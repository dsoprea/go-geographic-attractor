package geoattractorindex

import (
    "fmt"
    "os"
    "path"
    "reflect"
    "testing"

    "github.com/dsoprea/go-logging"

    // "github.com/dsoprea/go-geographic-attractor"
    "github.com/dsoprea/go-geographic-attractor/parse"
)

func getCityIndex(citydataFilepath string) *CityIndex {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)
            panic(err)
        }
    }()

    // Load countries.

    countrydataFilepath := path.Join(appPath, "test", "asset", "countryInfo.txt")

    f, err := os.Open(countrydataFilepath)
    log.PanicIf(err)

    defer f.Close()

    countries, err := geoattractorparse.BuildGeonamesCountryMapping(f)
    log.PanicIf(err)

    // Load cities.

    gp := geoattractorparse.NewGeonamesParser(countries)

    g, err := os.Open(citydataFilepath)
    log.PanicIf(err)

    defer g.Close()

    ci := NewCityIndex()

    err = ci.Load(gp, g)
    log.PanicIf(err)

    return ci
}

func TestCityIndex_Load(t *testing.T) {
    ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

    ls := ci.Stats()
    if ls.UnfilteredRecords != 100000 {
        t.Fatalf("The number of unfiltered records is not correct: (%d)", ls.UnfilteredRecords)
    } else if ls.RecordAdds != 281 {
        t.Fatalf("The number of added records is not correct: (%d)", ls.RecordAdds)
    } else if ls.RecordUpdates != 2 {
        t.Fatalf("The number of updated records is not correct: (%d)", ls.RecordUpdates)
    }
}

func dumpVisits(visits []VisitHistoryItem) {
    for _, vhi := range visits {
        fmt.Printf("%s: %s\n", vhi.Token, vhi.City)
    }
}

func TestCityIndex_Nearest_Miss(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)
            t.Fatalf("Panic.")
        }
    }()

    ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

    lasvegasCoordinates := []float64{36.175, -115.136389}

    _, _, _, err := ci.Nearest(lasvegasCoordinates[0], lasvegasCoordinates[1])
    if err == nil {
        t.Fatalf("Expected not-found error for Las Vegas (no error).")
    } else if log.Is(err, ErrNoNearestCity) == false {
        t.Fatalf("Expected not-found error for Las Vegas (error is not right type): [%s]", err)
    }
}

// TestCityIndex_Nearest_MultipleWithinSameCity just tests that multiple local
// points resolve to that same city. It's not profound.
func TestCityIndex_Nearest_MultipleWithinSameCity(t *testing.T) {
    ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

    anguillaCoordinates := []float64{18.21533, -63.02123}
    nearHotelCoordinates := []float64{18.216706, -63.020533}
    nearBusinessCoordinates := []float64{18.219121, -63.015356}

    sourceName1, _, cr1, err := ci.Nearest(anguillaCoordinates[0], anguillaCoordinates[1])
    log.PanicIf(err)

    if sourceName1 != "GeoNames" {
        t.Fatalf("Source-name for search (1) is not correct: [%s]", sourceName1)
    } else if cr1.Id != "3573511" {
        t.Fatalf("ID for search (1) is not correct: [%s]", cr1.Id)
    }

    sourceName2, _, cr2, err := ci.Nearest(nearHotelCoordinates[0], nearHotelCoordinates[1])
    log.PanicIf(err)

    if sourceName2 != sourceName1 {
        t.Fatalf("Source-name for search (2) is not correct: [%s]", sourceName2)
    } else if cr2.Id != cr1.Id {
        t.Fatalf("ID for search (2) is not correct: [%s]", cr2.Id)
    }

    sourceName3, _, cr3, err := ci.Nearest(nearBusinessCoordinates[0], nearBusinessCoordinates[1])
    log.PanicIf(err)

    if sourceName3 != sourceName1 {
        t.Fatalf("Source-name for search (3) is not correct: [%s]", sourceName3)
    } else if cr3.Id != cr1.Id {
        t.Fatalf("ID for search (3) is not correct: [%s]", cr3.Id)
    }
}

func TestCityIndex_Nearest_HitOnSmallAndAttractToLarge(t *testing.T) {
    ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

    clawsonCoordinates := []float64{42.53667, -83.15041}

    sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1])
    log.PanicIf(err)

    if sourceName != "GeoNames" {
        t.Fatalf("Source-name for search is not correct: [%s]", sourceName)
    } else if cr.Id != "4990752" {
        t.Fatalf("ID for search is not correct: [%s]", cr.Id)
    }

    if len(visits) != 24 {
        t.Fatalf("Number of visits not correct: (%d)", len(visits))
    }

    actual := make([]string, len(visits))
    for i, vhi := range visits {
        actual[i] = fmt.Sprintf("%s: %s", vhi.Token, vhi.City)
    }

    expected := []string{
        "8824c5ce97677fc5: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677fc4: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677fd: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677fc: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677f: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677c: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97677: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97674: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce9767: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce9764: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce977: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce974: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce97: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce94: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5ce9: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5cec: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5cf: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5cc: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5d: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>",
        "8824c5c: CityRecord<ID=[5007808] COUNTRY=[United States] CITY=[City of Royal Oak] POP=(57236) LAT=(42.5084000000) LON=(-83.1538700000)>",
        "8824c5: CityRecord<ID=[5007808] COUNTRY=[United States] CITY=[City of Royal Oak] POP=(57236) LAT=(42.5084000000) LON=(-83.1538700000)>",
        "8824c4: CityRecord<ID=[5012643] COUNTRY=[United States] CITY=[City of Troy] POP=(80980) LAT=(42.5817300000) LON=(-83.1457500000)>",
        "8824d: CityRecord<ID=[4990752] COUNTRY=[United States] CITY=[City of Detroit] POP=(713777) LAT=(42.3834100000) LON=(-83.1024100000)>",
        "8824c: CityRecord<ID=[4990752] COUNTRY=[United States] CITY=[City of Detroit] POP=(713777) LAT=(42.3834100000) LON=(-83.1024100000)>",
    }

    if reflect.DeepEqual(actual, expected) == false {
        t.Fatalf("Visit history not correct.")
    }
}

func TestCityIndex_Nearest_NearSmallAndNotNearLarge(t *testing.T) {
    ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

    trentonCoordinates := []float64{42.135582, -83.1928263}

    sourceName, visits, cr, err := ci.Nearest(trentonCoordinates[0], trentonCoordinates[1])
    log.PanicIf(err)

    // We should get Trenton in response (no large urban areas).
    if sourceName != "GeoNames" {
        t.Fatalf("Source-name for search is not correct: [%s]", sourceName)
    } else if cr.Id != "5012524" {
        t.Fatalf("ID for search is not correct: [%s]", cr.Id)
    }

    if len(visits) != 7 {
        t.Fatalf("Number of visits not correct: (%d)", len(visits))
    }

    actual := make([]string, len(visits))
    for i, vhi := range visits {
        actual[i] = fmt.Sprintf("%s: %s", vhi.Token, vhi.City)
    }

    expected := []string{
        "883b3914: CityRecord<ID=[5012524] COUNTRY=[United States] CITY=[City of Trenton] POP=(18853) LAT=(42.1394000000) LON=(-83.1930400000)>",
        "883b391: CityRecord<ID=[5012524] COUNTRY=[United States] CITY=[City of Trenton] POP=(18853) LAT=(42.1394000000) LON=(-83.1930400000)>",
        "883b394: CityRecord<ID=[5012524] COUNTRY=[United States] CITY=[City of Trenton] POP=(18853) LAT=(42.1394000000) LON=(-83.1930400000)>",
        "883b39: CityRecord<ID=[5012524] COUNTRY=[United States] CITY=[City of Trenton] POP=(18853) LAT=(42.1394000000) LON=(-83.1930400000)>",
        "883b3c: CityRecord<ID=[5012524] COUNTRY=[United States] CITY=[City of Trenton] POP=(18853) LAT=(42.1394000000) LON=(-83.1930400000)>",
        "883b3: CityRecord<ID=[4990516] COUNTRY=[United States] CITY=[City of Dearborn] POP=(98153) LAT=(42.3126900000) LON=(-83.2129400000)>",
        "883b4: CityRecord<ID=[4990516] COUNTRY=[United States] CITY=[City of Dearborn] POP=(98153) LAT=(42.3126900000) LON=(-83.2129400000)>",
    }

    if reflect.DeepEqual(actual, expected) == false {
        t.Fatalf("Visit history not correct.")
    }
}

func ExampleCityIndex_Nearest_AttractToLarge() {
    ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

    clawsonCoordinates := []float64{42.53667, -83.15041}

    sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1])
    log.PanicIf(err)

    for _, vhi := range visits {
        fmt.Printf("%s: %s\n", vhi.Token, vhi.City)
    }

    fmt.Printf("\n")

    fmt.Printf("Source: %s\n", sourceName)
    fmt.Printf("ID: %s\n", cr.Id)
    fmt.Printf("Country: %s\n", cr.Country)
    fmt.Printf("City: %s\n", cr.City)
    fmt.Printf("Population: %d\n", cr.Population)
    fmt.Printf("Latitude: %.10f\n", cr.Latitude)
    fmt.Printf("Longitude: %.10f\n", cr.Longitude)

    // Output:
    // 8824c5ce97677fc5: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677fc4: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677fd: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677fc: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677f: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677c: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97677: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97674: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce9767: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce9764: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce977: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce974: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce97: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce94: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5ce9: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5cec: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5cf: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5cc: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5d: CityRecord<ID=[4989009] COUNTRY=[United States] CITY=[City of Clawson] POP=(11825) LAT=(42.5366700000) LON=(-83.1504100000)>
    // 8824c5c: CityRecord<ID=[5007808] COUNTRY=[United States] CITY=[City of Royal Oak] POP=(57236) LAT=(42.5084000000) LON=(-83.1538700000)>
    // 8824c5: CityRecord<ID=[5007808] COUNTRY=[United States] CITY=[City of Royal Oak] POP=(57236) LAT=(42.5084000000) LON=(-83.1538700000)>
    // 8824c4: CityRecord<ID=[5012643] COUNTRY=[United States] CITY=[City of Troy] POP=(80980) LAT=(42.5817300000) LON=(-83.1457500000)>
    // 8824d: CityRecord<ID=[4990752] COUNTRY=[United States] CITY=[City of Detroit] POP=(713777) LAT=(42.3834100000) LON=(-83.1024100000)>
    // 8824c: CityRecord<ID=[4990752] COUNTRY=[United States] CITY=[City of Detroit] POP=(713777) LAT=(42.3834100000) LON=(-83.1024100000)>
    //
    // Source: GeoNames
    // ID: 4990752
    // Country: United States
    // City: City of Detroit
    // Population: 713777
    // Latitude: 42.3834100000
    // Longitude: -83.1024100000
}
