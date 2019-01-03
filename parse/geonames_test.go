package geoattractorparse

import (
    "os"
    "path"
    "reflect"
    "testing"
    "fmt"

    "github.com/dsoprea/go-geographic-attractor"
    "github.com/dsoprea/go-logging"
)

func getCountryMapping() map[string]string {
    filepath := path.Join(appPath, "test", "asset", "countryInfo.txt")

    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    countries, err := BuildGeonamesCountryMapping(f)
    log.PanicIf(err)

    return countries
}

func TestBuildGeonamesCountryMapping(t *testing.T) {
    countries := getCountryMapping()

    // Look for the first country in the list.

    name, found := countries["AD"]
    if found == false {
        t.Fatalf("Could not find country with acronym 'AD'.")
    } else if name != "Andorra" {
        t.Fatalf("The country found with acronym 'AD' was not correct.")
    }

    // Look for the last country in the list.

    name, found = countries["AN"]
    if found == false {
        t.Fatalf("Could not find country with acronym 'AN'.")
    } else if name != "Netherlands Antilles" {
        t.Fatalf("The country found with acronym 'AN' was not correct.")
    }

    // Look for a country from the middle of the list.

    name, found = countries["US"]
    if found == false {
        t.Fatalf("Could not find country with acronym 'US'.")
    } else if name != "United States" {
        t.Fatalf("The country found with acronym 'US' was not correct.")
    }
}

func TestGeonamesParser_Parse(t *testing.T) {
    countries := getCountryMapping()

    gp := NewGeonamesParser(countries)

    filepath := path.Join(testAssetsPath, "allCountries.txt.short")

    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    actual := make([]string, 0)
    cb := func(cr geoattractor.CityRecord) (err error) {
        actual = append(actual, cr.String())

        return nil
    }

    recordsCount, err := gp.Parse(f, cb)
    log.PanicIf(err)

    if recordsCount != 10000 {
        t.Fatalf("Number of records read is not correct: (%d)", recordsCount)
    }

    expected := []string{
        "CityRecord<ID=[3038999] COUNTRY=[Andorra] CITY=[Soldeu] POP=(602) LAT=(42.5768800000) LON=(1.6676900000)>",
        "CityRecord<ID=[3039154] COUNTRY=[Andorra] CITY=[El Tarter] POP=(1052) LAT=(42.5795200000) LON=(1.6536200000)>",
        "CityRecord<ID=[3039163] COUNTRY=[Andorra] CITY=[Sant Julià de Lòria] POP=(8022) LAT=(42.4637200000) LON=(1.4912900000)>",
        "CityRecord<ID=[3039604] COUNTRY=[Andorra] CITY=[Pas de la Casa] POP=(2363) LAT=(42.5427700000) LON=(1.7336100000)>",
        "CityRecord<ID=[3039678] COUNTRY=[Andorra] CITY=[Ordino] POP=(3066) LAT=(42.5562300000) LON=(1.5331900000)>",
        "CityRecord<ID=[3040051] COUNTRY=[Andorra] CITY=[les Escaldes] POP=(15853) LAT=(42.5072900000) LON=(1.5341400000)>",
        "CityRecord<ID=[3040132] COUNTRY=[Andorra] CITY=[la Massana] POP=(7211) LAT=(42.5449900000) LON=(1.5148300000)>",
        "CityRecord<ID=[3040140] COUNTRY=[Andorra] CITY=[l'Aldosa de canillo] POP=(195) LAT=(42.5789500000) LON=(1.6290200000)>",
        "CityRecord<ID=[3040141] COUNTRY=[Andorra] CITY=[l'Aldosa] POP=(594) LAT=(42.5439100000) LON=(1.5228900000)>",
        "CityRecord<ID=[3040686] COUNTRY=[Andorra] CITY=[Encamp] POP=(11223) LAT=(42.5347400000) LON=(1.5801400000)>",
        "CityRecord<ID=[3041204] COUNTRY=[Andorra] CITY=[Canillo] POP=(3292) LAT=(42.5676000000) LON=(1.5975600000)>",
        "CityRecord<ID=[3041519] COUNTRY=[Andorra] CITY=[Arinsal] POP=(1419) LAT=(42.5720500000) LON=(1.4845300000)>",
        "CityRecord<ID=[3041563] COUNTRY=[Andorra] CITY=[Andorra la Vella] POP=(20430) LAT=(42.5077900000) LON=(1.5210900000)>",
        "CityRecord<ID=[7302102] COUNTRY=[Andorra] CITY=[La Margineda] POP=(155) LAT=(42.4859900000) LON=(1.4902400000)>",
        "CityRecord<ID=[10630523] COUNTRY=[Andorra] CITY=[Puiol del Piu] POP=(400) LAT=(42.5652000000) LON=(1.4915900000)>",
        "CityRecord<ID=[290594] COUNTRY=[United Arab Emirates] CITY=[Umm al Qaywayn] POP=(44411) LAT=(25.5647300000) LON=(55.5551700000)>",
        "CityRecord<ID=[291074] COUNTRY=[United Arab Emirates] CITY=[Ras al-Khaimah] POP=(115949) LAT=(25.7895300000) LON=(55.9432000000)>",
        "CityRecord<ID=[291279] COUNTRY=[United Arab Emirates] CITY=[Muzayri‘] POP=(10000) LAT=(23.1435500000) LON=(53.7881000000)>",
        "CityRecord<ID=[291339] COUNTRY=[United Arab Emirates] CITY=[Murbaḩ] POP=(2000) LAT=(25.2762300000) LON=(56.3625600000)>",
        "CityRecord<ID=[291696] COUNTRY=[United Arab Emirates] CITY=[Khawr Fakkān] POP=(33575) LAT=(25.3313200000) LON=(56.3419900000)>",
        "CityRecord<ID=[292223] COUNTRY=[United Arab Emirates] CITY=[Dubai] POP=(1137347) LAT=(25.0657000000) LON=(55.1712800000)>",
        "CityRecord<ID=[292231] COUNTRY=[United Arab Emirates] CITY=[Dibba Al-Fujairah] POP=(30000) LAT=(25.5924600000) LON=(56.2617600000)>",
        "CityRecord<ID=[292239] COUNTRY=[United Arab Emirates] CITY=[Dibba Al-Hisn] POP=(26395) LAT=(25.6195500000) LON=(56.2729100000)>",
        "CityRecord<ID=[292672] COUNTRY=[United Arab Emirates] CITY=[Sharjah] POP=(543733) LAT=(25.3373700000) LON=(55.4120600000)>",
        "CityRecord<ID=[292688] COUNTRY=[United Arab Emirates] CITY=[Ar Ruways] POP=(16000) LAT=(24.1102800000) LON=(52.7305600000)>",
        "CityRecord<ID=[292878] COUNTRY=[United Arab Emirates] CITY=[Al Fujayrah] POP=(62415) LAT=(25.1164100000) LON=(56.3414100000)>",
        "CityRecord<ID=[292913] COUNTRY=[United Arab Emirates] CITY=[Al Ain] POP=(408733) LAT=(24.1916700000) LON=(55.7605600000)>",
        "CityRecord<ID=[292932] COUNTRY=[United Arab Emirates] CITY=[Ajman] POP=(226172) LAT=(25.4111100000) LON=(55.4350400000)>",
        "CityRecord<ID=[292953] COUNTRY=[United Arab Emirates] CITY=[Adh Dhayd] POP=(24716) LAT=(25.2881200000) LON=(55.8815700000)>",
        "CityRecord<ID=[292968] COUNTRY=[United Arab Emirates] CITY=[Abu Dhabi] POP=(603492) LAT=(24.4666700000) LON=(54.3666700000)>",
        "CityRecord<ID=[1120483] COUNTRY=[Afghanistan] CITY=[Kuhsān] POP=(12087) LAT=(34.6538900000) LON=(61.1977800000)>",
        "CityRecord<ID=[1120487] COUNTRY=[Afghanistan] CITY=[Tukzār] POP=(12021) LAT=(35.9483100000) LON=(66.4213200000)>",
        "CityRecord<ID=[1120711] COUNTRY=[Afghanistan] CITY=[Zindah Jān] POP=(10104) LAT=(34.3426400000) LON=(61.7467500000)>",
        "CityRecord<ID=[1120863] COUNTRY=[Afghanistan] CITY=[Zarghūn Shahr] POP=(13737) LAT=(32.8473400000) LON=(68.4457300000)>",
        "CityRecord<ID=[1120879] COUNTRY=[Afghanistan] CITY=[Zaṟah Sharan] POP=(7366) LAT=(33.1464100000) LON=(68.7921300000)>",
    }

    if reflect.DeepEqual(actual, expected) == false {
        for _, visit := range actual {
            fmt.Printf("%s\n", visit)
        }

        t.Fatalf("Results not expected.")
    }
}
