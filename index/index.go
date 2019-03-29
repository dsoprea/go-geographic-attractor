package geoattractorindex

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/dsoprea/go-logging"
	"github.com/kellydunn/golang-geo"
	"github.com/randomingenuity/go-utility/geographic"

	"github.com/dsoprea/go-geographic-attractor"
)

const (
	// Cache Nearest() lookups.
	MaxNearestLruEntries = 100
)

var (
	indexLogger = log.NewLogger("geoattractor.index")
)

var (
	ErrNoNearestCity = errors.New("no nearest city")
)

type indexEntry struct {
	CityRecord geoattractor.CityRecord
	Level      int

	// LeafCellId is the full cell-ID for the current city regardless of which
	// level we indexed it at.
	LeafCellToken string

	SourceName string
}

type AttractorStats struct {
	// UnfilteredRecords is the total number of records that were seen in the
	// file before any filtering was applied.
	UnfilteredRecords int `json:"unfiltered_records_parsed"`

	// RecordAdds are the number of new records committed to the index for new
	// cells/levels.
	RecordAdds int `json:"records_added_to_index"`

	// RecordUpdates are the number of records that replaced existing ones
	// (mutually exclusively with RecordAdds).
	RecordUpdates int `json:"records_updated_in_index"`

	// HaversineCalculations is how many times we've calculated distances
	// between points.
	HaversineCalculations int `json:"haversine_calculations"`

	CachedNearestHits   int
	CachedNearestMisses int
	CachedNearestShifts int
}

func (ls AttractorStats) String() string {
	return fmt.Sprintf("AttractorStats<UNFILTERED-RECORDS=(%d) ADDS=(%d) UPDATES=(%d) CACHE-HITS=(%d) CACHE-MISSES=(%d) CACHE-SHIFTS=(%d)>", ls.UnfilteredRecords, ls.RecordAdds, ls.RecordUpdates, ls.CachedNearestHits, ls.CachedNearestMisses, ls.CachedNearestShifts)
}

type cachedNearestInfo struct {
	sourceName string
	visits     []VisitHistoryItem
	cr         geoattractor.CityRecord
}

type CityIndex struct {
	index                   map[string][]*indexEntry
	stats                   AttractorStats
	urbanCentersEncountered map[string]geoattractor.CityRecord

	cachedNearest    map[string]cachedNearestInfo
	cachedNearestLru sort.StringSlice

	minimumSearchLevel           int
	urbanCenterMinimumPopulation int
}

// NewCityIndex returns a `CityIndex` instance. `minimumSearchLevel` specifies
// the smallest level (largest region) that we want to search for cities around
// a certain point.
func NewCityIndex(minimumSearchLevel int, urbanCenterMinimumPopulation int) *CityIndex {
	index := make(map[string][]*indexEntry)

	return &CityIndex{
		index:                   index,
		urbanCentersEncountered: make(map[string]geoattractor.CityRecord),

		cachedNearest:                make(map[string]cachedNearestInfo),
		cachedNearestLru:             make(sort.StringSlice, 0),
		minimumSearchLevel:           minimumSearchLevel,
		urbanCenterMinimumPopulation: urbanCenterMinimumPopulation,
	}
}

func (ci *CityIndex) Stats() AttractorStats {
	return ci.stats
}

// Load feeds the given city data into the index. Cities will be stored at
// multiple levels. If/when we experience collisions, we'll keep whichever has
// the larger population.
func (ci *CityIndex) Load(source geoattractor.CityRecordSource, r io.Reader, specificCityIds []string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	var cityIdsFilter sort.StringSlice
	if specificCityIds != nil {
		cityIdsFilter = sort.StringSlice(specificCityIds)
		cityIdsFilter.Sort()
	}

	cb := func(cr geoattractor.CityRecord) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
			}
		}()

		// Apply the filter.
		if cityIdsFilter != nil {
			i := cityIdsFilter.Search(cr.Id)
			if i >= len(cityIdsFilter) || cityIdsFilter[i] != cr.Id {
				return nil
			}
		}

		cellId := rigeo.S2CellFromCoordinates(cr.Latitude, cr.Longitude)
		token := cellId.ToToken()

		ie := &indexEntry{
			CityRecord: cr,
			Level:      cellId.Level(),

			// LeafCellId is the full cell-ID for the current city regardless of which
			// level we indexed it at.
			LeafCellToken: token,

			SourceName: source.Name(),
		}

		// Index this cell at all levels only to within the maximum area we'd
		// like to attract within. We assume that any area we visit will
		// hopefully be within this amount of distance from an urban center,
		// and, if not, at least one other city. Otherwise, that city won't be
		// matched within the index.
		for level := cellId.Level(); level >= ci.minimumSearchLevel; level-- {
			parentCellId := cellId.Parent(level)
			parentToken := parentCellId.ToToken()

			if existingEntries, found := ci.index[parentToken]; found == true {
				ci.index[parentToken] = append(existingEntries, ie)

				// Technically, this represents colocations.
				ci.stats.RecordUpdates++
			} else {
				// We either haven't seen this cell yet or the city we've
				// previously indexed is smaller than the current one.

				// TODO(dustin): !! Determine if preallocating additional slots would optimize. We might do this only for smaller levels in which the chance of collision increases.
				ci.index[parentToken] = []*indexEntry{ie}
				ci.stats.RecordAdds++
			}
		}

		return nil
	}

	recordsCount, err := source.Parse(r, cb)
	log.PanicIf(err)

	ci.stats.UnfilteredRecords = recordsCount

	return nil
}

type VisitHistoryItem struct {
	Token      string
	City       geoattractor.CityRecord
	SourceName string
}

// Nearest returns the nearest urban-center to the given coordinates, or, if
// none, the nearest city. Note that, because of how cells are layed out, some
// near urban centers won't be selected while others will be.
//
// Also returns the name of the data-source that produced the final result and
// the heirarchy of cities that surround the given coordinates up to the largest
// area that we index for urban centers in.
func (ci *CityIndex) Nearest(latitude, longitude float64, returnAllVisits bool) (sourceName string, visits []VisitHistoryItem, cr geoattractor.CityRecord, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	cellId := rigeo.S2CellFromCoordinates(latitude, longitude)

	// Use the cell-ID rather than the coordinates to key by (eliminates
	// precision and jitter issues).
	cacheKey := fmt.Sprintf("%d,%v", cellId, returnAllVisits)
	if cached, found := ci.cachedNearest[cacheKey]; found == true {
		ci.stats.CachedNearestHits++
		return cached.sourceName, cached.visits, cached.cr, nil
	} else {
		ci.stats.CachedNearestMisses++
	}

	// Efficiently collect all of the urban centers around our point using our
	// S2 index.

	if returnAllVisits == true {
		visits = make([]VisitHistoryItem, 0)
	}

	visitsUrbanCenters := make([]VisitHistoryItem, 0)
	nearestCities := make([]VisitHistoryItem, 0)
	for level := cellId.Level(); level >= ci.minimumSearchLevel; level-- {
		currentCellId := cellId.Parent(level)
		currentToken := currentCellId.ToToken()

		entries, found := ci.index[currentToken]
		if found == false {
			continue
		}

		// If this is our first hit on one (or more cities, if more than one is
		// very near).
		isNearestCities := len(nearestCities) == 0
		for _, ie := range entries {
			vhi := VisitHistoryItem{
				Token:      currentToken,
				City:       ie.CityRecord,
				SourceName: ie.SourceName,
			}

			if returnAllVisits == true {
				visits = append(visits, vhi)
			}

			if isNearestCities == true {
				nearestCities = append(nearestCities, vhi)
			}

			if int(ie.CityRecord.Population) >= ci.urbanCenterMinimumPopulation {
				visitsUrbanCenters = append(visitsUrbanCenters, vhi)

				ci.urbanCentersEncountered[ie.CityRecord.Id] = ie.CityRecord
			}
		}
	}

	// This will produce a more accurate result than S2 can on its own because
	// of how it cuts-up the world (e.g. we end-up not seeing cities or
	// grabbing cities further away before considering those that are nearer).

	var vhi VisitHistoryItem
	if len(visitsUrbanCenters) > 0 {
		vhi = ci.getNearestPoint(latitude, longitude, visitsUrbanCenters)
	} else {
		// If nothing else, just return the closest city found.

		// We don't actually have anything indexed for any of the cells
		// concentrically surrounding this location.
		if len(nearestCities) == 0 {
			log.Panic(ErrNoNearestCity)
		}

		vhi = ci.getNearestPoint(latitude, longitude, nearestCities)
	}

	cni := cachedNearestInfo{
		sourceName: vhi.SourceName,
		visits:     visits,
		cr:         vhi.City,
	}

	// Prune an entry out of the cache.

	if len(ci.cachedNearest) > MaxNearestLruEntries {
		oldestKey := ci.cachedNearestLru[0]
		ci.cachedNearestLru = ci.cachedNearestLru[1:]

		delete(ci.cachedNearest, oldestKey)

		ci.stats.CachedNearestShifts++
	}

	// Enroll in cache.

	if _, found := ci.cachedNearest[cacheKey]; found == true {
		i := ci.cachedNearestLru.Search(cacheKey)
		if i >= len(ci.cachedNearestLru) {
			log.Panicf("could not find existing cache entry in LRU")
		}

		if ci.cachedNearestLru[i] != cacheKey && i < len(ci.cachedNearestLru)-1 {
			// Move to end.
			ci.cachedNearestLru = append(ci.cachedNearestLru[:i], ci.cachedNearestLru[i+1], cacheKey)
		}
	} else {
		ci.cachedNearest[cacheKey] = cni

		ci.cachedNearestLru = append(ci.cachedNearestLru, cacheKey)
		ci.cachedNearestLru.Sort()
	}

	return vhi.SourceName, visits, vhi.City, nil
}

func (ci *CityIndex) UrbanCentersEncountered() map[string]geoattractor.CityRecord {
	return ci.urbanCentersEncountered
}

// getNearestPoint calculates the Haversine distance between the origin point
// and all points in the list and returns the nearest.
func (ci *CityIndex) getNearestPoint(originLatitude, originLongitude float64, queries []VisitHistoryItem) VisitHistoryItem {
	origin := geo.NewPoint(originLatitude, originLongitude)

	var closestDistance float64
	var closestVhi VisitHistoryItem

	empty := VisitHistoryItem{}

	for _, vhi := range queries {
		urbanP := geo.NewPoint(vhi.City.Latitude, vhi.City.Longitude)

		distance := origin.GreatCircleDistance(urbanP)
		ci.stats.HaversineCalculations++

		if closestVhi == empty || distance < closestDistance {
			closestDistance = distance
			closestVhi = vhi
		}
	}

	return closestVhi
}
