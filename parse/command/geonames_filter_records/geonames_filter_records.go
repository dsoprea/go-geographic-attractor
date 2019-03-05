package main

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path"
    "strings"

    "github.com/dsoprea/go-geographic-attractor"
    "github.com/dsoprea/go-geographic-attractor/parse"
    "github.com/dsoprea/go-logging"
    "github.com/jessevdk/go-flags"
)

var (
    arguments = new(parameters)
)

type parameters struct {
    CountryDataFilepath string   `long:"country-data-filepath" description:"GeoNames country-data file-path" required:"true"`
    InputFilepath       string   `long:"input-filepath" description:"GeoNames city- and population-data input file-path" required:"true"`
    OutputFilepath      string   `long:"output-filepath" description:"Output file-path"`
    JustCities          []string `long:"city" description:"Include city (looks like \"city,2-letter country\" or \"city\"); can be provided zero or more times"`
    MinimumPopulation   uint64   `long:"population" description:"Minimum population size"`
}

func main() {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)
            os.Exit(1)
        }
    }()

    p := flags.NewParser(arguments, flags.Default)

    _, err := p.Parse()
    if err != nil {
        os.Exit(1)
    }

    // Load countries.

    countrydataFile, err := os.Open(arguments.CountryDataFilepath)
    log.PanicIf(err)

    defer countrydataFile.Close()

    countries, err := geoattractorparse.BuildGeonamesCountryMapping(countrydataFile)
    log.PanicIf(err)

    // Load cities.

    gp := geoattractorparse.NewGeonamesParser(countries)

    var cityDataFile io.ReadCloser

    cityDataFilepath := arguments.InputFilepath

    if path.Ext(strings.ToLower(cityDataFilepath)) == ".zip" {
        zf, err := zip.OpenReader(cityDataFilepath)
        log.PanicIf(err)

        defer zf.Close()

        innerFilename := "allCountries.txt"
        for _, file := range zf.File {
            if file.Name == innerFilename {
                fc, err := file.Open()
                log.PanicIf(err)

                cityDataFile = fc
            }
        }

        if cityDataFile == nil {
            log.Panicf("Could not find file [%s] in the city-data archive: [%s]", innerFilename, cityDataFilepath)
        }
    } else {
        cityDataFile, err = os.Open(cityDataFilepath)
        log.PanicIf(err)
    }

    defer cityDataFile.Close()

    var outFile *os.File
    if arguments.OutputFilepath != "" {
        var err error
        outFile, err = os.Create(arguments.OutputFilepath)
        log.PanicIf(err)
    } else {
        outFile = os.Stdout
    }

    cityAndCountryFilter := make([]string, 0)
    cityFilter := make([]string, 0)
    if arguments.JustCities != nil {
        for _, phrase := range arguments.JustCities {
            parts := strings.Split(strings.ToLower(phrase), ",")

            len_ := len(parts)
            if len_ == 2 {
                cityAndCountryFilter = append(cityAndCountryFilter, fmt.Sprintf("%s,%s", parts[0], parts[1]))
            } else if len_ == 1 {
                cityFilter = append(cityFilter, parts[0])
            } else {
                log.Panicf("city filter phrase not formatted correctly: [%s]", phrase)
            }
        }

        // TODO(dustin): Sort filters. Search properly.
    }

    predicate := func(cr *geoattractor.CityRecord) (keep bool, err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
            }
        }()

        if arguments.MinimumPopulation > 0 && cr.Population < arguments.MinimumPopulation {
            return false, nil
        }

        country := strings.ToLower(cr.Country)
        city := strings.ToLower(cr.City)

        key := fmt.Sprintf("%s,%s", country, city)
        for _, filter := range cityAndCountryFilter {
            if filter == key {
                return true, nil
            }
        }

        for _, filter := range cityFilter {
            if filter == city {
                return true, nil
            }
        }

        return false, nil
    }

    err = gp.Filter(cityDataFile, outFile, predicate)
    log.PanicIf(err)

    if arguments.OutputFilepath != "" {
        outFile.Close()
    }
}
