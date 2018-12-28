package geoattractorparser

import (
    "os"
    "path"
    "testing"

    "github.com/dsoprea/go-logging"
)

func TestBuildGeonamesCountryMapping(t *testing.T) {
    filepath := path.Join(testAssetsPath, "countryInfo.txt")

    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    countries, err := BuildGeonamesCountryMapping(f)
    log.PanicIf(err)

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
