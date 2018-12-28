package geoattractorparser

import (
    "encoding/csv"
    "fmt"
    "io"
    "strconv"

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

        if len(record) > 0 && len(record[0]) > 0 && record[0][0] == '#' {
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

type CityCb func(cr geoattractor.CityRecord) (err error)

func (gp *GeonamesParser) Parse(r io.Reader, cityCb CityCb) (err error) {
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

        geonamesId := record[0]
        name := record[1]
        latitudeRaw := record[4]
        longitudeRaw := record[5]
        featureClass := record[6]
        featureCode := record[7]
        countryCode := record[8]
        populationRaw := record[14]

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

        // If we get here, we have a tangible population value.

        countryName, found := gp.countries[countryCode]
        if found == false {
            log.Panicf("could not resolve country with code [%s]", countryCode)
        }

        latitude, err := strconv.ParseFloat(latitudeRaw, 64)
        log.PanicIf(err)

        longitude, err := strconv.ParseFloat(longitudeRaw, 64)
        log.PanicIf(err)

        cr := geoattractor.CityRecord{
            Country:    countryName,
            City:       name,
            Population: population,
            Latitude:   latitude,
            Longitude:  longitude,
        }

        fmt.Printf("CITY: [%s]\n", cr.City)
        fmt.Printf("COUNTRY: [%s]\n", cr.Country)
        fmt.Printf("FEATURE: CLASS=[%s] CODE=[%s]\n", featureClass, featureCode)
        fmt.Printf("\n")

        err = cityCb(cr)
        log.PanicIf(err)
    }

    return nil
}
