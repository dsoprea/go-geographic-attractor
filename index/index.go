package geoattractorindex

import (
    "errors"
    "fmt"
    "io"
    "sort"

    "github.com/dsoprea/go-logging"
    "github.com/golang/geo/s2"

    "github.com/dsoprea/go-geographic-attractor"
)

const (
    // MinimumLevelForUrbanCenterAttraction is the lowest level that we'll
    // compile the city with the highest population within.
    MinimumLevelForUrbanCenterAttraction = 7

    // UrbanCenterMinimumPopulation is the minimum population a city requires in
    // order to be considered an urban/metropolitan center.
    UrbanCenterMinimumPopulation = 100000
)

var (
    indexLogger = log.NewLogger("geoattractor.index")
)

var (
    ErrNoNearestCity = errors.New("no nearest city")
)

type indexEntry struct {
    Info  geoattractor.CityRecord
    Level int

    // LeafCellId is the full cell-ID for the current city regardless of which
    // level we indexed it at.
    LeafCellToken string

    SourceName string
}

type LoadStats struct {
    // UnfilteredRecords is the total number of records that were seen in the
    // file before any filtering was applied.
    UnfilteredRecords int `json:"unfiltered_records_parsed"`

    // RecordAdds are the number of new records committed to the index for new
    // cells/levels.
    RecordAdds int `json:"records_added_to_index"`

    // RecordUpdates are the number of records that replaced existing ones
    // (mutually exclusively with RecordAdds).
    RecordUpdates int `json:"records_updated_in_index"`
}

func (ls LoadStats) String() string {
    return fmt.Sprintf("LoadStats<UNFILTERED-RECORDS=(%d) ADDS=(%d) UPDATES=(%d)>", ls.UnfilteredRecords, ls.RecordAdds, ls.RecordUpdates)
}

type CityIndex struct {
    index map[string]*indexEntry
    stats LoadStats
}

func NewCityIndex() *CityIndex {
    index := make(map[string]*indexEntry)

    return &CityIndex{
        index: index,
    }
}

func (ci *CityIndex) Stats() LoadStats {
    return ci.stats
}

// Load feeds the given city data into the index. Cities will be stored at
// multiple levels. If/when we experience collisions, we'll keep whichever has
// the larger population.
func (ci *CityIndex) Load(source geoattractor.CityRecordSource, r io.Reader) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    cb := func(cr geoattractor.CityRecord) (err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
            }
        }()

        ll := s2.LatLngFromDegrees(cr.Latitude, cr.Longitude)
        if ll.IsValid() == false {
            indexLogger.Warningf(nil, "Coordinates for city [%s] with ID [%s] from source [%s] are not valid: (%.10f), (%.10f)", cr.City, cr.Id, source.Name(), cr.Latitude, cr.Longitude)
            return nil
        }

        cellId := s2.CellIDFromLatLng(ll)
        if cellId.IsValid() == false {
            indexLogger.Warningf(nil, "Cell [%s] generated for city [%s] with ID [%s] from source [%s] is not valid.", uint64(cellId), cr.City, cr.Id, source.Name())
            return nil
        }

        token := cellId.ToToken()

        ie := &indexEntry{
            Info:  cr,
            Level: cellId.Level(),

            // LeafCellId is the full cell-ID for the current city regardless of which
            // level we indexed it at.
            LeafCellToken: cellId.ToToken(),

            SourceName: source.Name(),
        }

        // Store the leaf (full resolution) entry.
        ci.index[token] = ie

        // Index this cell at all levels only to within the maximum area we'd
        // like to attract within. We assume that any area we visit will
        // hopefully be within this amount of distance from an urban center,
        // and, if not, at least one other city. Otherwise, that city won't be
        // matched within the index.
        for level := cellId.Level() - 1; level >= MinimumLevelForUrbanCenterAttraction; level-- {
            parentCellId := cellId.Parent(level)
            parentToken := parentCellId.ToToken()

            existingEntry, found := ci.index[parentToken]

            if found == false || existingEntry.Info.Population < cr.Population {
                // We either haven't seen this cell yet or the city we've
                // previously indexed is smaller than the current one.

                ci.index[parentToken] = ie

                if found == true {
                    ci.stats.RecordUpdates++
                } else {
                    ci.stats.RecordAdds++
                }
            }
        }

        return nil
    }

    recordsCount, err := source.Parse(r, cb)
    log.PanicIf(err)

    ci.stats.UnfilteredRecords = recordsCount

    final := make([]string, len(ci.index))
    i := 0
    for token, _ := range ci.index {
        final[i] = token
        i++
    }

    ss := sort.StringSlice(final)
    ss.Sort()

    return nil
}

type VisitHistoryItem struct {
    Token string
    City  geoattractor.CityRecord
}

// Nearest returns the nearest urban-center to the given coordinates, or, if
// none, the nearest city. Note that, because of how cells are layed out, some
// near urban centers won't be selected while others will be.
//
// Also returns the name of the data-source that produced the final result and
// the heirarchy of cities that surround the given coordinates up to the largest
// area that we index for urban centers in.
func (ci *CityIndex) Nearest(latitude, longitude float64) (sourceName string, visits []VisitHistoryItem, cr geoattractor.CityRecord, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ll := s2.LatLngFromDegrees(latitude, longitude)
    if ll.IsValid() == false {
        log.Panicf("Coordinates not valid: (%.10f), (%.10f)", latitude, longitude)
    }

    cellId := s2.CellIDFromLatLng(ll)
    if cellId.IsValid() == false {
        log.Panicf("Determined cell not valid for coordinates: (%.10f), (%.10f)", latitude, longitude)
    }

    // Successfully search for matches at progressively larger areas.

    var firstMatch *indexEntry
    var firstUrbanCenter *indexEntry
    visits = make([]VisitHistoryItem, 0)
    for level := cellId.Level(); level >= MinimumLevelForUrbanCenterAttraction; level-- {
        currentCellId := cellId.Parent(level)

        currentToken := currentCellId.ToToken()

        ie, found := ci.index[currentToken]
        if found == false {
            continue
        }

        vhi := VisitHistoryItem{
            Token: currentToken,
            City:  ie.Info,
        }

        visits = append(visits, vhi)

        if firstMatch == nil {
            firstMatch = ie
        }

        // Keep track of the first city we encounter that we regard as a urban center.
        if ie.Info.Population > UrbanCenterMinimumPopulation && firstUrbanCenter == nil {
            firstUrbanCenter = ie
        }
    }

    // We don't actually have anything indexed for any of the cells
    // concentrically surrounding this location.
    if firstMatch == nil && firstUrbanCenter == nil {
        log.Panic(ErrNoNearestCity)
    }

    // Return the first urban-center that we encountered as we visited the
    // concentric cells (if any).
    if firstUrbanCenter != nil && firstUrbanCenter.Info.Population >= UrbanCenterMinimumPopulation {
        return firstUrbanCenter.SourceName, visits, firstUrbanCenter.Info, nil
    }

    // If nothing else, just return the city found closest to our location.
    return firstMatch.SourceName, visits, firstMatch.Info, nil
}
