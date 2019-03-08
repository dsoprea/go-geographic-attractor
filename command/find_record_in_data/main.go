package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dsoprea/go-logging"
	"github.com/jessevdk/go-flags"
	"github.com/randomingenuity/go-utility/geographic"

	"github.com/dsoprea/go-geographic-attractor"
	"github.com/dsoprea/go-geographic-attractor/index"
	"github.com/dsoprea/go-geographic-attractor/parse"
)

type parameters struct {
	CountryDataFilepath string   `short:"c" long:"country-data-filepath" description:"GeoNames country-data file-path"`
	CityDataFilepath    string   `short:"p" long:"city-data-filepath" description:"GeoNames city- and population-data file-path"`
	IdList              []string `short:"i" long:"record-id" description:"ID of record to find (can be provided zero or more times)"`
	CoordinatesList     []string `short:"C" long:"coordinates" description:"Exact latitude/longitude to search (e.g. '12.345,67.891'; can be provided zero or more times)"`
	OnlyUrbanCenters    bool     `short:"u" long:"urban-centers" description:"Only print urban centers"`
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

	cellsList := make([]uint64, len(arguments.CoordinatesList))

	for i, coordinatePhrase := range arguments.CoordinatesList {
		parts := strings.Split(coordinatePhrase, ",")
		if len(parts) != 2 {
			log.Panicf("coordinate phrase is not exactly two parts: [%s]", coordinatePhrase)
		}

		latitute, err := strconv.ParseFloat(parts[0], 64)
		log.PanicIf(err)

		longitude, err := strconv.ParseFloat(parts[1], 64)
		log.PanicIf(err)

		cell := rigeo.S2CellFromCoordinates(latitute, longitude)
		cellsList[i] = uint64(cell)
	}

	hasQualifiers := len(arguments.IdList) > 0 || len(cellsList) > 0

	cb := func(cr geoattractor.CityRecord) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
			}
		}()

		// The parser implementation is expected to filter by everything but
		// population.
		if arguments.OnlyUrbanCenters == true && cr.Population < geoattractorindex.UrbanCenterMinimumPopulation {
			return nil
		}

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
			currentCell := rigeo.S2CellFromCoordinates(cr.Latitude, cr.Longitude)
			currentCellId := uint64(currentCell)

			for _, cellId := range cellsList {
				if cellId == currentCellId {
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
