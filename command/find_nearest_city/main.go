package main

import (
    "archive/zip"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path"
    "strings"

    "github.com/dsoprea/go-logging"
    "github.com/jessevdk/go-flags"

    "github.com/dsoprea/go-geographic-attractor/index"
    "github.com/dsoprea/go-geographic-attractor/parse"
)

type parameters struct {
    CountryDataFilepath string  `short:"c" long:"country-data-filepath" description:"GeoNames country-data file-path"`
    CityDataFilepath    string  `short:"p" long:"city-data-filepath" description:"GeoNames city- and population-data file-path"`
    Latitude            float64 `short:"a" long:"latitude" description:"Latitude" required:"true"`
    Longitude           float64 `short:"o" long:"longitude" description:"Longitude" required:"true"`
    Verbose             bool    `short:"v" long:"verbose" description:"Print logging"`
    Json                bool    `short:"j" long:"json" description:"Print as JSON"`
}

var (
    arguments = new(parameters)
)

var (
    commandLogger = log.NewLogger("command/find_nearest_city")
)

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

    countryDataFilepath := arguments.CountryDataFilepath
    if countryDataFilepath == "" {
        countryDataFilepath = os.Getenv("GGA_COUNTRY_DATA_FILEPATH")

        if countryDataFilepath == "" {
            fmt.Printf("Country-data file-path not provided or defined via GGA_COUNTRY_DATA_FILEPATH.\n")
            os.Exit(2)
        }
    }

    cityDataFilepath := arguments.CityDataFilepath
    if cityDataFilepath == "" {
        cityDataFilepath = os.Getenv("GGA_CITY_DATA_FILEPATH")

        if cityDataFilepath == "" {
            fmt.Printf("City-data file-path not provided or defined via GGA_CITY_DATA_FILEPATH.\n")
            os.Exit(2)
        }
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

    ci := geoattractorindex.NewCityIndex()

    err = ci.Load(gp, cityDataFile)
    log.PanicIf(err)

    sourceName, visits, cr, err := ci.Nearest(arguments.Latitude, arguments.Longitude)
    if err != nil {
        if log.Is(err, geoattractorindex.ErrNoNearestCity) == true {
            fmt.Printf("No nearest city found.\n")
            os.Exit(10)
        }

        log.Panic(err)
    }

    if arguments.Json == true {
        result := map[string]interface{}{
            "Result": cr,
            "Stats":  ci.Stats(),
        }

        encoded, err := json.MarshalIndent(result, "", "  ")
        log.PanicIf(err)

        fmt.Println(string(encoded))
    } else {
        if arguments.Verbose == true {
            for i, vhi := range visits {
                fmt.Printf("VISIT(% 2d): %s: %s\n", i, vhi.Token, vhi.City)
            }

            fmt.Printf("\n")
        }

        fmt.Printf("Source: %s\n", sourceName)
        fmt.Printf("ID: %s\n", cr.Id)
        fmt.Printf("Country: %s\n", cr.Country)
        fmt.Printf("City: %s\n", cr.City)
        fmt.Printf("Population: %d\n", cr.Population)
        fmt.Printf("Latitude: %.10f\n", cr.Latitude)
        fmt.Printf("Longitude: %.10f\n", cr.Longitude)
    }
}
