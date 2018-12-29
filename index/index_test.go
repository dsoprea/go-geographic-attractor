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
    ci := getCityIndex(path.Join(appPath, "test", "asset", "countryInfo.txt"))

    ls := ci.Stats()
    if ls.UnfilteredRecords != 100000 {
        t.Fatalf("The number of unfiltered records is not correct: (%d)", ls.UnfilteredRecords)
    } else if ls.RecordAdds != 3511 {
        t.Fatalf("The number of added records is not correct: (%d)", ls.RecordAdds)
    } else if ls.RecordUpdates != 67 {
        t.Fatalf("The number of updated records is not correct: (%d)", ls.RecordUpdates)
    }
}

// TODO(dustin): !! Debug this.

// func TestCityIndex_Nearest_WithinSameCity(t *testing.T) {
//     ci := getCityIndex(path.Join(appPath, "test", "asset", "countryInfo.txt"))

//     anguillaCoordinates := []float64{18.21533, -63.02123}
//     nearHotelCoordinates := []float64{18.216706, -63.020533}
//     nearBusinessCoordinates := []float64{18.219121, -63.015356}

//     sourceName1, cr1, err := ci.Nearest(anguillaCoordinates[0], anguillaCoordinates[1])
//     log.PanicIf(err)

//     if sourceName1 != "GeoNames" {
//         t.Fatalf("Source-name for search (1) is not correct: [%s]", sourceName1)
//     } else if cr1.Id != "11205444" {
//         t.Fatalf("ID for search (1) is not correct: [%s]", cr1.Id)
//     }

//     sourceName2, cr2, err := ci.Nearest(nearHotelCoordinates[0], nearHotelCoordinates[1])
//     log.PanicIf(err)

//     if sourceName2 != sourceName1 {
//         t.Fatalf("Source-name for search (2) is not correct: [%s]", sourceName2)
//     } else if cr2.Id != cr1.Id {
//         t.Fatalf("ID for search (2) is not correct: [%s]", cr2.Id)
//     }

//     sourceName3, cr3, err := ci.Nearest(nearBusinessCoordinates[0], nearBusinessCoordinates[1])
//     log.PanicIf(err)

//     if sourceName3 != sourceName1 {
//         t.Fatalf("Source-name for search (3) is not correct: [%s]", sourceName3)
//     } else if cr3.Id != cr1.Id {
//         t.Fatalf("ID for search (3) is not correct: [%s]", cr3.Id)
//     }
// }

func TestCityIndex_Nearest_NearUrbanArea_One(t *testing.T) {
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

// func TestCityIndex_Nearest_NearUrbanArea_Many(t *testing.T) {
//     ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

//     // uniques := make(map[string]geoattractor.CityRecord)
//     // for _, ie := range ci.index {
//     //     uniques[ie.LeafCellToken] = ie.Info
//     // }

//     // for _, cr := range uniques {
//     //     if cr.Population > 100000 {
//     //         fmt.Printf("%s\n", cr)
//     //     }
//     // }

//     clawsonCoordinates := []float64{42.533333, -83.146389}

//     // nearHotelCoordinates := []float64{18.216706, -63.020533}
//     // nearBusinessCoordinates := []float64{18.219121, -63.015356}

//     sourceName1, visits, cr1, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1])
//     log.PanicIf(err)

//     for _, vh := range visits {
//         fmt.Printf("%s\n", vh)
//     }

//     fmt.Printf("\n")

//     // NOTE(dustin): !! Why isn't Warren being visited prior to Detroit?

//     if sourceName1 != "GeoNames" {
//         t.Fatalf("Source-name for search (1) is not correct: [%s]", sourceName1)
//         // } else if cr1.Id != "11205444" {
//         //     t.Fatalf("ID for search (1) is not correct: [%s]", cr1.Id)
//     }

//     fmt.Printf("%s\n", cr1)

//     // sourceName2, cr2, err := ci.Nearest(nearHotelCoordinates[0], nearHotelCoordinates[1])
//     // log.PanicIf(err)

//     // if sourceName2 != sourceName1 {
//     //     t.Fatalf("Source-name for search (2) is not correct: [%s]", sourceName2)
//     // } else if cr2.Id != cr1.Id {
//     //     t.Fatalf("ID for search (2) is not correct: [%s]", cr2.Id)
//     // }

//     // sourceName3, cr3, err := ci.Nearest(nearBusinessCoordinates[0], nearBusinessCoordinates[1])
//     // log.PanicIf(err)

//     // if sourceName3 != sourceName1 {
//     //     t.Fatalf("Source-name for search (3) is not correct: [%s]", sourceName3)
//     // } else if cr3.Id != cr1.Id {
//     //     t.Fatalf("ID for search (3) is not correct: [%s]", cr3.Id)
//     // }
// }
