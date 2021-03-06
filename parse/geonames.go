package geoattractorparse

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"archive/zip"
	"encoding/csv"

	"github.com/dsoprea/go-geographic-attractor"
	"github.com/dsoprea/go-logging"
)

// BuildGeonamesCountryMapping parses the GeoNames countryInfo.txt file.
func BuildGeonamesCountryMapping(r io.Reader) (countries map[string]string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	c := csv.NewReader(r)
	c.Comma = '\t'

	countries = make(map[string]string)
	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}

		if len(record) != 19 {
			// A line that doesn't look like a record.

			continue
		} else if record[0][0] == '#' {
			// A commented line that was somehow interpreted as a record.

			continue
		}

		acronym := record[0]
		name := record[4]

		countries[acronym] = name
	}

	return countries, nil
}

type GeonamesParser struct {
	countries map[string]string
}

func NewGeonamesParser(countries map[string]string) *GeonamesParser {
	return &GeonamesParser{
		countries: countries,
	}
}

func (gp *GeonamesParser) Parse(r io.Reader, cityRecordCb geoattractor.CityRecordCb) (recordsCount int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	c := csv.NewReader(r)
	c.Comma = '\t'

	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}

		// From http://download.geonames.org/export/dump:
		//
		//  0: geonameid         : integer id of record in geonames database
		//  1: name              : name of geographical point (utf8) varchar(200)
		//  2: asciiname         : name of geographical point in plain ascii characters, varchar(200)
		//  3: alternatenames    : alternatenames, comma separated, ascii names automatically transliterated, convenience attribute from alternatename table, varchar(10000)
		//  4: latitude          : latitude in decimal degrees (wgs84)
		//  5: longitude         : longitude in decimal degrees (wgs84)
		//  6: feature class     : see http://www.geonames.org/export/codes.html, char(1)
		//  7: feature code      : see http://www.geonames.org/export/codes.html, varchar(10)
		//  8: country code      : ISO-3166 2-letter country code, 2 characters
		//  9: cc2               : alternate country codes, comma separated, ISO-3166 2-letter country code, 200 characters
		// 10: admin1 code       : fipscode (subject to change to iso code), see exceptions below, see file admin1Codes.txt for display names of this code; varchar(20)
		// 11: admin2 code       : code for the second administrative division, a county in the US, see file admin2Codes.txt; varchar(80)
		// 12: admin3 code       : code for third level administrative division, varchar(20)
		// 13: admin4 code       : code for fourth level administrative division, varchar(20)
		// 14: population        : bigint (8 byte int)
		// 15: elevation         : in meters, integer
		// 16: dem               : digital elevation model, srtm3 or gtopo30, average elevation of 3''x3'' (ca 90mx90m) or 30''x30'' (ca 900mx900m) area in meters, integer. srtm processed by cgiar/ciat.
		// 17: timezone          : the iana timezone id (see file timeZone.txt) varchar(40)
		// 18: modification date : date of last modification in yyyy-MM-dd format

		if len(record) != 19 {
			// A line that doesn't look like a record.

			continue
		} else if record[0][0] == '#' {
			// A commented line that was somehow interpreted as a record.

			continue
		}

		geonamesId := record[0]

		// We've accidentally fed-in the country-list by accident so many times
		// that now we're just protecting against it.
		_, err = strconv.ParseUint(geonamesId, 10, 64)
		if err != nil {
			log.Panicf("first column doesn't look like an integer; are we looking at the right kind of file? %s", record)
		}

		name := record[1]
		latitudeRaw := record[4]
		longitudeRaw := record[5]
		featureClass := record[6]
		featureCode := record[7]
		countryCode := record[8]
		admin1Code := record[10]
		populationRaw := record[14]

		// In the case (name == "Commonwealth of Independent States").
		if countryCode == "" {
			continue
		}

		dumpRecord := func() {
			fmt.Printf("\n")
			fmt.Printf("RECORD\n")
			fmt.Printf("======\n")

			for i, part := range record {
				fmt.Printf("%02d: [%s] (%d)\n", i, part, len(part))
			}

			fmt.Printf("\n")
		}

		// TODO(dustin): !! Move these out to a configurable filter.

		if featureClass != "P" {
			continue
		} else if featureCode != "PPLC" && strings.HasPrefix(featureCode, "PPLA") == false && featureCode != "PPL" && featureCode != "PPLX" && featureCode != "PPLL" {
			// Filter for any populated place type. These all appear to depend on
			// the size of the place and no particular classification applies.

			continue
		}

		if populationRaw == "" || populationRaw == "null" {
			continue
		} else if name == "" {
			log.Panicf("no city name found for GeoNames ID [%s]", geonamesId)
		}

		population, err := strconv.ParseUint(populationRaw, 10, 64)
		log.PanicIf(err)

		if population == 0 {
			continue
		}

		recordsCount++

		if cityRecordCb != nil {
			// If we get here, we have a tangible population value.

			countryName, found := gp.countries[countryCode]
			if found == false {
				dumpRecord()

				log.Panicf("could not resolve country with code [%s] ((%d) countries known)", countryCode, len(gp.countries))
			}

			latitude, err := strconv.ParseFloat(latitudeRaw, 64)
			log.PanicIf(err)

			longitude, err := strconv.ParseFloat(longitudeRaw, 64)
			log.PanicIf(err)

			cr := geoattractor.CityRecord{
				Id:            geonamesId,
				Country:       countryName,
				ProvinceState: admin1Code,
				City:          name,
				Population:    population,
				Latitude:      latitude,
				Longitude:     longitude,
			}

			err = cityRecordCb(cr)
			log.PanicIf(err)
		}
	}

	return recordsCount, nil
}

func (gp *GeonamesParser) Name() (name string) {
	return "GeoNames"
}

func NewGeonamesParserWithFiles(countryDataFilepath string) (gp *GeonamesParser, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if countryDataFilepath == "" {
		countryDataFilepath = os.Getenv("GGA_COUNTRY_DATA_FILEPATH")

		if countryDataFilepath == "" {
			log.Panicf("country-data file-path not provided or defined via GGA_COUNTRY_DATA_FILEPATH")
		}
	}

	// Load countries.

	countrydataFile, err := os.Open(countryDataFilepath)
	log.PanicIf(err)

	defer countrydataFile.Close()

	countries, err := BuildGeonamesCountryMapping(countrydataFile)
	log.PanicIf(err)

	// Load cities.

	gp = NewGeonamesParser(countries)
	return gp, nil
}

func GetCitydataReadCloser(cityDataFilepath string) (rc io.ReadCloser, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if cityDataFilepath == "" {
		cityDataFilepath = os.Getenv("GGA_CITY_DATA_FILEPATH")

		if cityDataFilepath == "" {
			log.Panicf("city-data file-path not provided or defined via GGA_CITY_DATA_FILEPATH")
		}
	}

	if path.Ext(strings.ToLower(cityDataFilepath)) == ".zip" {
		zf, err := zip.OpenReader(cityDataFilepath)
		log.PanicIf(err)

		defer zf.Close()

		innerFilename := "allCountries.txt"
		for _, file := range zf.File {
			if file.Name == innerFilename {
				rc, err = file.Open()
				log.PanicIf(err)
			}
		}

		if rc == nil {
			log.Panicf("Could not find file [%s] in the city-data archive: [%s]", innerFilename, cityDataFilepath)
		}
	} else {
		rc, err = os.Open(cityDataFilepath)
		log.PanicIf(err)
	}

	return rc, nil
}
