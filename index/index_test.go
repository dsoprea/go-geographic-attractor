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

func getCityIndex(cityDataFilepath string) *CityIndex {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			panic(err)
		}
	}()

	// Load countries.

	countryDataFilepath := path.Join(appPath, "test", "asset", "countryInfo.txt")

	f, err := os.Open(countryDataFilepath)
	log.PanicIf(err)

	defer f.Close()

	countries, err := geoattractorparse.BuildGeonamesCountryMapping(f)
	log.PanicIf(err)

	// Load cities.

	gp := geoattractorparse.NewGeonamesParser(countries)

	g, err := os.Open(cityDataFilepath)
	log.PanicIf(err)

	defer g.Close()

	ci := NewCityIndex()

	err = ci.Load(gp, g)
	log.PanicIf(err)

	return ci
}

func TestCityIndex_Load(t *testing.T) {
	ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

	ls := ci.Stats()
	if ls.RecordAdds != 12176 {
		t.Fatalf("The number of added records is not correct: (%d)", ls.RecordAdds)
	} else if ls.RecordUpdates != 361 {
		t.Fatalf("The number of updated records is not correct: (%d)", ls.RecordUpdates)
	}
}

func dumpVisits(visits []VisitHistoryItem) {
	for _, vhi := range visits {
		fmt.Printf("%s: %s\n", vhi.Token, vhi.City)
	}
}

func TestCityIndex_Nearest_Miss(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Panic.")
		}
	}()

	ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

	lasvegasCoordinates := []float64{36.175, -115.136389}

	_, _, _, err := ci.Nearest(lasvegasCoordinates[0], lasvegasCoordinates[1])
	if err == nil {
		t.Fatalf("Expected not-found error for Las Vegas (no error).")
	} else if log.Is(err, ErrNoNearestCity) == false {
		t.Fatalf("Expected not-found error for Las Vegas (error is not right type): [%s]", err)
	}
}

// TestCityIndex_Nearest_MultipleWithinSameCity just tests that multiple local
// points resolve to that same city. It's not profound.
func TestCityIndex_Nearest_MultipleWithinSameCity(t *testing.T) {
	ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

	alainCoordinates := []float64{24.1916700000, 55.7605600000}
	alburaimiCoordinates := []float64{24.269806, 55.831959}
	omanCoordinates := []float64{24.032976, 56.116184}

	sourceName1, _, cr1, err := ci.Nearest(alainCoordinates[0], alainCoordinates[1])
	log.PanicIf(err)

	if sourceName1 != "GeoNames" {
		t.Fatalf("Source-name for search (1) is not correct: [%s]", sourceName1)
	} else if cr1.Id != "292913" {
		t.Fatalf("ID for search (1) is not correct: [%s]", cr1.Id)
	}

	sourceName2, _, cr2, err := ci.Nearest(alburaimiCoordinates[0], alburaimiCoordinates[1])
	log.PanicIf(err)

	if sourceName2 != sourceName1 {
		t.Fatalf("Source-name for search (2) is not correct: [%s]", sourceName2)
	} else if cr2.Id != cr1.Id {
		t.Fatalf("ID for search (2) is not correct: [%s] != [%s]", cr2.Id, cr1.Id)
	}

	sourceName3, _, cr3, err := ci.Nearest(omanCoordinates[0], omanCoordinates[1])
	log.PanicIf(err)

	if sourceName3 != sourceName1 {
		t.Fatalf("Source-name for search (3) is not correct: [%s]", sourceName3)
	} else if cr3.Id != cr1.Id {
		t.Fatalf("ID for search (3) is not correct: [%s] != [%s]", cr3.Id, cr1.Id)
	}
}

func TestCityIndex_Nearest_HitOnSmallAndAttractToLarge(t *testing.T) {
	ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

	clawsonCoordinates := []float64{42.53667, -83.15041}

	sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1])
	log.PanicIf(err)

	if sourceName != "GeoNames" {
		t.Fatalf("Source-name for search is not correct: [%s]", sourceName)
	} else if cr.Id != "5014051" {
		t.Fatalf("ID for search is not correct: [%s]", cr.Id)
	}

	actual := make([]string, len(visits))
	for i, vhi := range visits {
		actual[i] = fmt.Sprintf("%s: %s", vhi.Token, vhi.City)
	}

	expected := []string{
		"8824c5cc: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5d: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5c: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c4: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824d: CityRecord<ID=[5014051] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000) S2=[8824d0a18dc66fa9]>",
		"8824c: CityRecord<ID=[5014051] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000) S2=[8824d0a18dc66fa9]>",
	}

	if reflect.DeepEqual(actual, expected) == false {
		for _, visit := range actual {
			fmt.Printf("%s\n", visit)
		}

		t.Fatalf("Visit history not correct.")
	}
}

func TestCityIndex_Nearest_NearSmallAndNotNearLarge(t *testing.T) {
	ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

	hillsdaleCoordinates := []float64{41.9275396, -84.6694791}

	sourceName, visits, cr, err := ci.Nearest(hillsdaleCoordinates[0], hillsdaleCoordinates[1])
	log.PanicIf(err)

	// We should get Trenton in response (no large urban areas).
	if sourceName != "GeoNames" {
		t.Fatalf("Source-name for search is not correct: [%s]", sourceName)
	} else if cr.Id != "4996107" {
		t.Fatalf("ID for search is not correct: [%s]", cr.Id)
	}

	actual := make([]string, len(visits))
	for i, vhi := range visits {
		actual[i] = fmt.Sprintf("%s: %s", vhi.Token, vhi.City)
	}

	expected := []string{
		"883d73: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d74: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d7: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d4: CityRecord<ID=[4997384] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Jackson] POP=(33133) LAT=(42.2458700000) LON=(-84.4013500000) S2=[883d257732f676b9]>",
	}

	if reflect.DeepEqual(actual, expected) == false {
		for _, visit := range actual {
			fmt.Printf("%s\n", visit)
		}

		t.Fatalf("Visit history not correct.")
	}
}

func ExampleCityIndex_Nearest() {
	// Load countries.

	countryDataFilepath := path.Join(appPath, "test", "asset", "countryInfo.txt")

	f, err := os.Open(countryDataFilepath)
	log.PanicIf(err)

	defer f.Close()

	countries, err := geoattractorparse.BuildGeonamesCountryMapping(f)
	log.PanicIf(err)

	// Load cities.

	gp := geoattractorparse.NewGeonamesParser(countries)

	cityDataFilepath := path.Join(appPath, "index", "test", "asset", "allCountries.txt.detroit_area_handpicked")
	g, err := os.Open(cityDataFilepath)
	log.PanicIf(err)

	defer g.Close()

	ci := NewCityIndex()

	err = ci.Load(gp, g)
	log.PanicIf(err)

	// Do the query.

	clawsonCoordinates := []float64{42.53667, -83.15041}

	sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1])
	log.PanicIf(err)

	// Print the results.

	for _, vhi := range visits {
		fmt.Printf("%s: %s\n", vhi.Token, vhi.City)
	}

	fmt.Printf("\n")

	fmt.Printf("Source: %s\n", sourceName)
	fmt.Printf("ID: %s\n", cr.Id)
	fmt.Printf("Country: %s\n", cr.Country)
	fmt.Printf("City: %s\n", cr.City)
	fmt.Printf("Population: %d\n", cr.Population)
	fmt.Printf("Latitude: %.10f\n", cr.Latitude)
	fmt.Printf("Longitude: %.10f\n", cr.Longitude)

	// Output:
	// 8824c5cc: CityRecord<ID=[4989005] COUNTRY=[United States] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000)>
	// 8824c5d: CityRecord<ID=[4989005] COUNTRY=[United States] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000)>
	// 8824c5c: CityRecord<ID=[4989005] COUNTRY=[United States] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000)>
	// 8824c5: CityRecord<ID=[4989005] COUNTRY=[United States] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000)>
	// 8824c4: CityRecord<ID=[5012639] COUNTRY=[United States] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000)>
	// 8824d: CityRecord<ID=[5014051] COUNTRY=[United States] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000)>
	// 8824c: CityRecord<ID=[5014051] COUNTRY=[United States] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000)>

	// Source: GeoNames
	// ID: 5014051
	// Country: United States
	// City: Warren
	// Population: 134056
	// Latitude: 42.4904400000
	// Longitude: -83.0130400000
}
