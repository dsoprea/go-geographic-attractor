package geoattractorparse

import (
    "os"
    "path"
    "reflect"
    "testing"

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

    defer f.Close()

    actual := make([]string, 0)
    cb := func(cr geoattractor.CityRecord) (err error) {
        actual = append(actual, cr.String())

        return nil
    }

    recordsCount, err := gp.Parse(f, cb)
    log.PanicIf(err)

    if recordsCount != 1000 {
        t.Fatalf("Number of records read is not correct: (%d)", recordsCount)
    }

    expected := []string{
        "CityRecord<ID=[3039162] COUNTRY=[Andorra] CITY=[Sant Julià de Lòria] POP=(9448) LAT=(42.4624700000) LON=(1.4824700000)>",
        "CityRecord<ID=[3039676] COUNTRY=[Andorra] CITY=[Ordino] POP=(3467) LAT=(42.5975800000) LON=(1.5257300000)>",
    }

    if reflect.DeepEqual(actual, expected) == false {
        t.Fatalf("Results not expected:\n%v", actual)
    }
}
