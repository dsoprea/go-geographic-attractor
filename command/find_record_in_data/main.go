package main

import (
	"fmt"
	"os"

	"github.com/dsoprea/go-logging"
	"github.com/jessevdk/go-flags"

	"github.com/dsoprea/go-geographic-attractor"
	"github.com/dsoprea/go-geographic-attractor/parse"
)

type parameters struct {
	CountryDataFilepath string   `short:"c" long:"country-data-filepath" description:"GeoNames country-data file-path"`
	CityDataFilepath    string   `short:"p" long:"city-data-filepath" description:"GeoNames city- and population-data file-path"`
	IdList              []string `short:"i" long:"record-id" description:"ID of record to find (can be provided zero or more times)"`
	CoordinatesList     []string `short:"C" long:"coordinates" description:"Exact latitude/longitude to search (e.g. '12.345,67.891'; can be provided zero or more times)"`
}

var (
	arguments = new(parameters)
)

var (
	commandLogger = log.NewLogger("command/find_record_in_data")
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
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	gp, err := geoattractorparse.NewGeonamesParserWithFiles(arguments.CountryDataFilepath)
	log.PanicIf(err)

	cityDataReadcloser, err := geoattractorparse.GetCitydataReadCloser(arguments.CityDataFilepath)
	log.PanicIf(err)

	defer cityDataReadcloser.Close()

	hasQualifiers := arguments.IdList != nil || arguments.CoordinatesList != nil

	cb := func(cr geoattractor.CityRecord) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
			}
		}()

		hit := false
		if arguments.IdList != nil {
			for _, idRaw := range arguments.IdList {
				if idRaw == cr.Id {
					hit = true
					break
				}
			}
		}

		if hit == false && arguments.CoordinatesList != nil {
			currentCoordinates := fmt.Sprintf("%.6f,%.6f", cr.Latitude, cr.Longitude)

			for _, coordinates := range arguments.CoordinatesList {
				if coordinates == currentCoordinates {
					hit = true
					break
				}
			}
		}

		// If no qualifiers are given, just print all of the records. This is
		// for exploring the data since the acceptance criteria can make it
		// difficult to the naked eye to know which records will be indexed.
		if hasQualifiers == true && hit == false {
			return nil
		}

		fmt.Printf("%s\n", cr)

		return nil
	}

	recordsCount, err := gp.Parse(cityDataReadcloser, cb)
	log.PanicIf(err)

	fmt.Printf("(%d) records scanned.\n", recordsCount)
}
