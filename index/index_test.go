package geoattractorindex

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-geographic-attractor"
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

	ci := NewTestCityIndex()

	err = ci.Load(gp, g, nil)
	log.PanicIf(err)

	return ci
}

func TestCityIndex_Load(t *testing.T) {
	ci := getCityIndex(path.Join(appPath, "index", "test", "asset", "allCountries.txt.head"))

	ls := ci.Stats()
	if ls.RecordAdds != 12776 {
		t.Fatalf("The number of added records is not correct: (%d)", ls.RecordAdds)
	} else if ls.RecordUpdates != 1624 {
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

	_, _, _, err := ci.Nearest(lasvegasCoordinates[0], lasvegasCoordinates[1], false)
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

	sourceName1, _, cr1, err := ci.Nearest(alainCoordinates[0], alainCoordinates[1], false)
	log.PanicIf(err)

	if sourceName1 != "GeoNames" {
		t.Fatalf("Source-name for search (1) is not correct: [%s]", sourceName1)
	} else if cr1.Id != "292913" {
		t.Fatalf("ID for search (1) is not correct: [%s]", cr1.Id)
	}

	sourceName2, _, cr2, err := ci.Nearest(alburaimiCoordinates[0], alburaimiCoordinates[1], false)
	log.PanicIf(err)

	if sourceName2 != sourceName1 {
		t.Fatalf("Source-name for search (2) is not correct: [%s]", sourceName2)
	} else if cr2.Id != cr1.Id {
		t.Fatalf("ID for search (2) is not correct: [%s] != [%s]", cr2.Id, cr1.Id)
	}

	sourceName3, _, cr3, err := ci.Nearest(omanCoordinates[0], omanCoordinates[1], false)
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

	sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1], true)
	log.PanicIf(err)

	if sourceName != "GeoNames" {
		t.Fatalf("Source-name for search is not correct: [%s]", sourceName)
	} else if cr.Id != "5011148" {
		t.Fatalf("ID for search is not correct: [%s]", cr.Id)
	}

	actual := make([]string, len(visits))
	for i, vhi := range visits {
		actual[i] = fmt.Sprintf("%s: %s", vhi.Token, vhi.City)
	}

	expected := []string{
		"8824c5cc: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5cc: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5d: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5d: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5c: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5c: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c5: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c4: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824c4: CityRecord<ID=[4985891] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Beverly Hills] POP=(10267) LAT=(42.5239200000) LON=(-83.2232600000) S2=[8824c7c6cfdc46d7]>",
		"8824c4: CityRecord<ID=[4986172] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Birmingham] POP=(20857) LAT=(42.5467000000) LON=(-83.2113200000) S2=[8824c700917ab677]>",
		"8824c4: CityRecord<ID=[4986429] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Bloomfield Hills] POP=(4004) LAT=(42.5836400000) LON=(-83.2454900000) S2=[8824c75476e7ce3b]>",
		"8824c4: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c4: CityRecord<ID=[5007402] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Rochester Hills] POP=(73424) LAT=(42.6583700000) LON=(-83.1499300000) S2=[8824c209e91136a5]>",
		"8824c4: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824c4: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824d: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824d: CityRecord<ID=[4985744] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Berkley] POP=(15268) LAT=(42.5030900000) LON=(-83.1835400000) S2=[8824c8a0f8f0fbdf]>",
		"8824d: CityRecord<ID=[4985891] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Beverly Hills] POP=(10267) LAT=(42.5239200000) LON=(-83.2232600000) S2=[8824c7c6cfdc46d7]>",
		"8824d: CityRecord<ID=[4986172] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Birmingham] POP=(20857) LAT=(42.5467000000) LON=(-83.2113200000) S2=[8824c700917ab677]>",
		"8824d: CityRecord<ID=[4986429] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Bloomfield Hills] POP=(4004) LAT=(42.5836400000) LON=(-83.2454900000) S2=[8824c75476e7ce3b]>",
		"8824d: CityRecord<ID=[4988400] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Center Line] POP=(8320) LAT=(42.4850400000) LON=(-83.0277000000) S2=[8824d08f9be326ad]>",
		"8824d: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824d: CityRecord<ID=[4989133] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clinton Township] POP=(99753) LAT=(42.5869800000) LON=(-82.9199200000) S2=[8824df6828a8d70d]>",
		"8824d: CityRecord<ID=[4991735] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Eastpointe] POP=(32657) LAT=(42.4683700000) LON=(-82.9554700000) S2=[8824d70b8a1bfd85]>",
		"8824d: CityRecord<ID=[4992635] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Ferndale] POP=(20177) LAT=(42.4605900000) LON=(-83.1346500000) S2=[8824cee31d69a0a3]>",
		"8824d: CityRecord<ID=[4993369] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Fraser] POP=(14636) LAT=(42.5392000000) LON=(-82.9493700000) S2=[8824d8fd4f96562b]>",
		"8824d: CityRecord<ID=[4994862] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Grosse Pointe] POP=(5232) LAT=(42.3861500000) LON=(-82.9118600000) S2=[8824d58bfcb6289d]>",
		"8824d: CityRecord<ID=[4994868] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Grosse Pointe Park] POP=(11220) LAT=(42.3758700000) LON=(-82.9374200000) S2=[8824d5a64ff86b33]>",
		"8824d: CityRecord<ID=[4995197] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hamtramck] POP=(22002) LAT=(42.3928200000) LON=(-83.0496400000) S2=[8824d2444ca18dbb]>",
		"8824d: CityRecord<ID=[4995368] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Harper Woods] POP=(13836) LAT=(42.4330900000) LON=(-82.9240800000) S2=[8824d63e628c12ef]>",
		"8824d: CityRecord<ID=[4995664] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hazel Park] POP=(16597) LAT=(42.4625400000) LON=(-83.1040900000) S2=[8824cfb4e9ca2e4f]>",
		"8824d: CityRecord<ID=[4996017] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Highland Park] POP=(10949) LAT=(42.4055900000) LON=(-83.0968700000) S2=[8824cdebb932fabd]>",
		"8824d: CityRecord<ID=[4996832] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Huntington Woods] POP=(6340) LAT=(42.4805900000) LON=(-83.1668700000) S2=[8824c8cb6e07e689]>",
		"8824d: CityRecord<ID=[4998900] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Lathrup Village] POP=(4135) LAT=(42.4964200000) LON=(-83.2227100000) S2=[8824c86e6b32379f]>",
		"8824d: CityRecord<ID=[5000500] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Madison Heights] POP=(30198) LAT=(42.4858700000) LON=(-83.1052000000) S2=[8824cf93f79aa573]>",
		"8824d: CityRecord<ID=[5004188] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Oak Park] POP=(29752) LAT=(42.4594800000) LON=(-83.1827100000) S2=[8824c923bbec1675]>",
		"8824d: CityRecord<ID=[5006011] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Pleasant Ridge] POP=(2556) LAT=(42.4711500000) LON=(-83.1421500000) S2=[8824cf1860ae671b]>",
		"8824d: CityRecord<ID=[5007402] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Rochester Hills] POP=(73424) LAT=(42.6583700000) LON=(-83.1499300000) S2=[8824c209e91136a5]>",
		"8824d: CityRecord<ID=[5007655] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Roseville] POP=(47637) LAT=(42.4972600000) LON=(-82.9371400000) S2=[8824d825a93fd115]>",
		"8824d: CityRecord<ID=[5007804] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Royal Oak] POP=(59008) LAT=(42.4894800000) LON=(-83.1446500000) S2=[8824cf426afeadeb]>",
		"8824d: CityRecord<ID=[5010636] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Southfield] POP=(73156) LAT=(42.4733700000) LON=(-83.2218700000) S2=[8824c84e9cc57315]>",
		"8824d: CityRecord<ID=[5011148] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Sterling Heights] POP=(132052) LAT=(42.5803100000) LON=(-83.0302000000) S2=[8824dc7b8dc14d09]>",
		"8824d: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824d: CityRecord<ID=[5013061] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Utica] POP=(4942) LAT=(42.6261400000) LON=(-83.0335400000) S2=[8824dda71da1f19b]>",
		"8824d: CityRecord<ID=[5014051] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000) S2=[8824d0a18dc66fa9]>",
		"8824d: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824c: CityRecord<ID=[4984067] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Almont] POP=(2723) LAT=(42.9205800000) LON=(-83.0449300000) S2=[8824f9addab97707]>",
		"8824c: CityRecord<ID=[4984565] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Auburn Hills] POP=(22672) LAT=(42.6875300000) LON=(-83.2341000000) S2=[8824eac14916a94d]>",
		"8824c: CityRecord<ID=[4985744] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Berkley] POP=(15268) LAT=(42.5030900000) LON=(-83.1835400000) S2=[8824c8a0f8f0fbdf]>",
		"8824c: CityRecord<ID=[4985891] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Beverly Hills] POP=(10267) LAT=(42.5239200000) LON=(-83.2232600000) S2=[8824c7c6cfdc46d7]>",
		"8824c: CityRecord<ID=[4986099] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Bingham Farms] POP=(1133) LAT=(42.5158700000) LON=(-83.2732600000) S2=[8824b8221822a2df]>",
		"8824c: CityRecord<ID=[4986172] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Birmingham] POP=(20857) LAT=(42.5467000000) LON=(-83.2113200000) S2=[8824c700917ab677]>",
		"8824c: CityRecord<ID=[4986429] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Bloomfield Hills] POP=(4004) LAT=(42.5836400000) LON=(-83.2454900000) S2=[8824c75476e7ce3b]>",
		"8824c: CityRecord<ID=[4988400] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Center Line] POP=(8320) LAT=(42.4850400000) LON=(-83.0277000000) S2=[8824d08f9be326ad]>",
		"8824c: CityRecord<ID=[4988997] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clarkston] POP=(1035) LAT=(42.7358600000) LON=(-83.4188300000) S2=[8824975eff9d9513]>",
		"8824c: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
		"8824c: CityRecord<ID=[4989133] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clinton Township] POP=(99753) LAT=(42.5869800000) LON=(-82.9199200000) S2=[8824df6828a8d70d]>",
		"8824c: CityRecord<ID=[4991218] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Dryden] POP=(941) LAT=(42.9461400000) LON=(-83.1238300000) S2=[8824f64b1bacaa75]>",
		"8824c: CityRecord<ID=[4991735] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Eastpointe] POP=(32657) LAT=(42.4683700000) LON=(-82.9554700000) S2=[8824d70b8a1bfd85]>",
		"8824c: CityRecord<ID=[4992519] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Farmington] POP=(10523) LAT=(42.4644800000) LON=(-83.3763200000) S2=[8824b109a7e2f587]>",
		"8824c: CityRecord<ID=[4992523] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Farmington Hills] POP=(81330) LAT=(42.4853100000) LON=(-83.3771600000) S2=[8824b0f85b33478b]>",
		"8824c: CityRecord<ID=[4992635] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Ferndale] POP=(20177) LAT=(42.4605900000) LON=(-83.1346500000) S2=[8824cee31d69a0a3]>",
		"8824c: CityRecord<ID=[4993335] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Franklin] POP=(3237) LAT=(42.5222600000) LON=(-83.3060400000) S2=[8824b9c53261a953]>",
		"8824c: CityRecord<ID=[4993369] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Fraser] POP=(14636) LAT=(42.5392000000) LON=(-82.9493700000) S2=[8824d8fd4f96562b]>",
		"8824c: CityRecord<ID=[4994154] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Goodrich] POP=(1831) LAT=(42.9169700000) LON=(-83.5063400000) S2=[882486ab47d9c425]>",
		"8824c: CityRecord<ID=[4994862] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Grosse Pointe] POP=(5232) LAT=(42.3861500000) LON=(-82.9118600000) S2=[8824d58bfcb6289d]>",
		"8824c: CityRecord<ID=[4994868] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Grosse Pointe Park] POP=(11220) LAT=(42.3758700000) LON=(-82.9374200000) S2=[8824d5a64ff86b33]>",
		"8824c: CityRecord<ID=[4995197] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hamtramck] POP=(22002) LAT=(42.3928200000) LON=(-83.0496400000) S2=[8824d2444ca18dbb]>",
		"8824c: CityRecord<ID=[4995368] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Harper Woods] POP=(13836) LAT=(42.4330900000) LON=(-82.9240800000) S2=[8824d63e628c12ef]>",
		"8824c: CityRecord<ID=[4995664] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hazel Park] POP=(16597) LAT=(42.4625400000) LON=(-83.1040900000) S2=[8824cfb4e9ca2e4f]>",
		"8824c: CityRecord<ID=[4996017] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Highland Park] POP=(10949) LAT=(42.4055900000) LON=(-83.0968700000) S2=[8824cdebb932fabd]>",
		"8824c: CityRecord<ID=[4996832] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Huntington Woods] POP=(6340) LAT=(42.4805900000) LON=(-83.1668700000) S2=[8824c8cb6e07e689]>",
		"8824c: CityRecord<ID=[4997868] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Keego Harbor] POP=(3029) LAT=(42.6080900000) LON=(-83.3438200000) S2=[8824bc1971399d75]>",
		"8824c: CityRecord<ID=[4998516] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Lake Angelus] POP=(297) LAT=(42.6986400000) LON=(-83.3166000000) S2=[882495c871182f65]>",
		"8824c: CityRecord<ID=[4998587] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Lake Orion] POP=(3051) LAT=(42.7844800000) LON=(-83.2396600000) S2=[8824ed2626f43ec7]>",
		"8824c: CityRecord<ID=[4998900] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Lathrup Village] POP=(4135) LAT=(42.4964200000) LON=(-83.2227100000) S2=[8824c86e6b32379f]>",
		"8824c: CityRecord<ID=[4999097] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Leonard] POP=(410) LAT=(42.8653100000) LON=(-83.1427100000) S2=[8824f112d80b63f7]>",
		"8824c: CityRecord<ID=[4999837] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Livonia] POP=(94635) LAT=(42.3683700000) LON=(-83.3527100000) S2=[8824b346025e4699]>",
		"8824c: CityRecord<ID=[5000500] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Madison Heights] POP=(30198) LAT=(42.4858700000) LON=(-83.1052000000) S2=[8824cf93f79aa573]>",
		"8824c: CityRecord<ID=[5001755] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Metamora] POP=(566) LAT=(42.9414200000) LON=(-83.2891100000) S2=[88248afe6662e88f]>",
		"8824c: CityRecord<ID=[5003956] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Northville] POP=(6010) LAT=(42.4311500000) LON=(-83.4832700000) S2=[8824ac3db472df9b]>",
		"8824c: CityRecord<ID=[5004062] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Novi] POP=(58723) LAT=(42.4805900000) LON=(-83.4754900000) S2=[8824aee2d62f3ac1]>",
		"8824c: CityRecord<ID=[5004188] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Oak Park] POP=(29752) LAT=(42.4594800000) LON=(-83.1827100000) S2=[8824c923bbec1675]>",
		"8824c: CityRecord<ID=[5004551] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Orchard Lake] POP=(2245) LAT=(42.5830900000) LON=(-83.3593800000) S2=[8824bb9290f37025]>",
		"8824c: CityRecord<ID=[5004593] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Ortonville] POP=(1463) LAT=(42.8522500000) LON=(-83.4430000000) S2=[8824856ab34c64a3]>",
		"8824c: CityRecord<ID=[5004817] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Oxford] POP=(3534) LAT=(42.8247500000) LON=(-83.2646600000) S2=[88248d5910271d1d]>",
		"8824c: CityRecord<ID=[5006011] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Pleasant Ridge] POP=(2556) LAT=(42.4711500000) LON=(-83.1421500000) S2=[8824cf1860ae671b]>",
		"8824c: CityRecord<ID=[5006059] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Plymouth] POP=(8905) LAT=(42.3714300000) LON=(-83.4702100000) S2=[8824acd137cbfcdb]>",
		"8824c: CityRecord<ID=[5006166] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Pontiac] POP=(59917) LAT=(42.6389200000) LON=(-83.2910500000) S2=[8824bfb28601e2d3]>",
		"8824c: CityRecord<ID=[5006917] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Redford] POP=(49936) LAT=(42.3833700000) LON=(-83.2966000000) S2=[8824b4e3187e95eb]>",
		"8824c: CityRecord<ID=[5007400] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Rochester] POP=(12993) LAT=(42.6805900000) LON=(-83.1338200000) S2=[8824e9b9c8dfe7f7]>",
		"8824c: CityRecord<ID=[5007402] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Rochester Hills] POP=(73424) LAT=(42.6583700000) LON=(-83.1499300000) S2=[8824c209e91136a5]>",
		"8824c: CityRecord<ID=[5007525] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Romeo] POP=(3625) LAT=(42.8028100000) LON=(-83.0129900000) S2=[8824e4b5c0b4364d]>",
		"8824c: CityRecord<ID=[5007655] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Roseville] POP=(47637) LAT=(42.4972600000) LON=(-82.9371400000) S2=[8824d825a93fd115]>",
		"8824c: CityRecord<ID=[5007804] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Royal Oak] POP=(59008) LAT=(42.4894800000) LON=(-83.1446500000) S2=[8824cf426afeadeb]>",
		"8824c: CityRecord<ID=[5009586] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Shelby] POP=(74099) LAT=(42.6708700000) LON=(-83.0329800000) S2=[8824e774da50f5c3]>",
		"8824c: CityRecord<ID=[5010636] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Southfield] POP=(73156) LAT=(42.4733700000) LON=(-83.2218700000) S2=[8824c84e9cc57315]>",
		"8824c: CityRecord<ID=[5011148] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Sterling Heights] POP=(132052) LAT=(42.5803100000) LON=(-83.0302000000) S2=[8824dc7b8dc14d09]>",
		"8824c: CityRecord<ID=[5011761] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Sylvan Lake] POP=(1785) LAT=(42.6114200000) LON=(-83.3285500000) S2=[8824be9883cd97db]>",
		"8824c: CityRecord<ID=[5012639] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Troy] POP=(83280) LAT=(42.6055900000) LON=(-83.1499300000) S2=[8824c3c40a768751]>",
		"8824c: CityRecord<ID=[5013061] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Utica] POP=(4942) LAT=(42.6261400000) LON=(-83.0335400000) S2=[8824dda71da1f19b]>",
		"8824c: CityRecord<ID=[5013961] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Walled Lake] POP=(7110) LAT=(42.5378100000) LON=(-83.4810500000) S2=[8824a594107a3e33]>",
		"8824c: CityRecord<ID=[5014051] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Warren] POP=(134056) LAT=(42.4904400000) LON=(-83.0130400000) S2=[8824d0a18dc66fa9]>",
		"8824c: CityRecord<ID=[5014130] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Waterford] POP=(75737) LAT=(42.6930300000) LON=(-83.4118100000) S2=[882497f2bcf5da7d]>",
		"8824c: CityRecord<ID=[5015351] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Wixom] POP=(13746) LAT=(42.5247600000) LON=(-83.5363300000) S2=[8824a89d9d87dce1]>",
		"8824c: CityRecord<ID=[5015416] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Wolverine Lake] POP=(4312) LAT=(42.5567000000) LON=(-83.4738300000) S2=[8824a5ade7b1c3f5]>",
		"8824c: CityRecord<ID=[7259621] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[West Bloomfield Township] POP=(64690) LAT=(42.5689100000) LON=(-83.3835600000) S2=[8824bb093a6f74c9]>",
		"8824c: CityRecord<ID=[4989005] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Clawson] POP=(12015) LAT=(42.5333700000) LON=(-83.1463200000) S2=[8824c5c88b28e955]>",
	}

	if reflect.DeepEqual(actual, expected) == false {
		for _, visit := range actual {
			fmt.Printf("\"%s\",\n", visit)
		}

		t.Fatalf("Visit history not correct.")
	}
}

func TestCityIndex_Nearest_NearSmallAndNotNearLarge(t *testing.T) {
	ci := getCityIndex(path.Join(testAssetsPath, "allCountries.txt.detroit_area_handpicked"))

	hillsdaleCoordinates := []float64{41.9275396, -84.6694791}

	sourceName, visits, cr, err := ci.Nearest(hillsdaleCoordinates[0], hillsdaleCoordinates[1], true)
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
		"883d74: CityRecord<ID=[4983996] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Allen] POP=(189) LAT=(41.9569900000) LON=(-84.7677400000) S2=[883d70022c11ecad]>",
		"883d74: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d74: CityRecord<ID=[5006848] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Reading] POP=(1056) LAT=(41.8394900000) LON=(-84.7480100000) S2=[883d77a0a78666d5]>",
		"883d7: CityRecord<ID=[4983996] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Allen] POP=(189) LAT=(41.9569900000) LON=(-84.7677400000) S2=[883d70022c11ecad]>",
		"883d7: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d7: CityRecord<ID=[4997698] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Jonesville] POP=(2220) LAT=(41.9842100000) LON=(-84.6619000000) S2=[883d6db09c39293b]>",
		"883d7: CityRecord<ID=[4999410] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Litchfield] POP=(1347) LAT=(42.0439300000) LON=(-84.7574600000) S2=[883d68b6b58fdc23]>",
		"883d7: CityRecord<ID=[5006647] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Quincy] POP=(1640) LAT=(41.9442100000) LON=(-84.8838500000) S2=[883d7ce9e6841a53]>",
		"883d7: CityRecord<ID=[5006848] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Reading] POP=(1056) LAT=(41.8394900000) LON=(-84.7480100000) S2=[883d77a0a78666d5]>",
		"883d4: CityRecord<ID=[4983802] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Addison] POP=(594) LAT=(41.9864300000) LON=(-84.3471700000) S2=[883d1c7a4c4fd4d3]>",
		"883d4: CityRecord<ID=[4983905] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Albion] POP=(8229) LAT=(42.2431000000) LON=(-84.7530300000) S2=[883d452299ea62f7]>",
		"883d4: CityRecord<ID=[4983996] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Allen] POP=(189) LAT=(41.9569900000) LON=(-84.7677400000) S2=[883d70022c11ecad]>",
		"883d4: CityRecord<ID=[4988380] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Cement City] POP=(426) LAT=(42.0700400000) LON=(-84.3305000000) S2=[883d1f16f7dcaef5]>",
		"883d4: CityRecord<ID=[4989442] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Concord] POP=(1050) LAT=(42.1778200000) LON=(-84.6430200000) S2=[883d40d978d59873]>",
		"883d4: CityRecord<ID=[4995249] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hanover] POP=(431) LAT=(42.1011500000) LON=(-84.5519000000) S2=[883d158f3dad081d]>",
		"883d4: CityRecord<ID=[4996107] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hillsdale] POP=(8163) LAT=(41.9200500000) LON=(-84.6305100000) S2=[883d72e6ee142c29]>",
		"883d4: CityRecord<ID=[4996369] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Homer] POP=(1630) LAT=(42.1458800000) LON=(-84.8088600000) S2=[883d5d0edea99b73]>",
		"883d4: CityRecord<ID=[4996718] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Hudson] POP=(2241) LAT=(41.8550500000) LON=(-84.3538400000) S2=[883d01535179f561]>",
		"883d4: CityRecord<ID=[4997384] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Jackson] POP=(33133) LAT=(42.2458700000) LON=(-84.4013500000) S2=[883d257732f676b9]>",
		"883d4: CityRecord<ID=[4997698] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Jonesville] POP=(2220) LAT=(41.9842100000) LON=(-84.6619000000) S2=[883d6db09c39293b]>",
		"883d4: CityRecord<ID=[4999410] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Litchfield] POP=(1347) LAT=(42.0439300000) LON=(-84.7574600000) S2=[883d68b6b58fdc23]>",
		"883d4: CityRecord<ID=[5001813] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Michigan Center] POP=(4672) LAT=(42.2330900000) LON=(-84.3271800000) S2=[883d26474636f58f]>",
		"883d4: CityRecord<ID=[5003589] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[North Adams] POP=(472) LAT=(41.9708800000) LON=(-84.5257800000) S2=[883d12211192a33d]>",
		"883d4: CityRecord<ID=[5005034] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Parma] POP=(760) LAT=(42.2583700000) LON=(-84.5996900000) S2=[883d3805c1d135c9]>",
		"883d4: CityRecord<ID=[5006647] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Quincy] POP=(1640) LAT=(41.9442100000) LON=(-84.8838500000) S2=[883d7ce9e6841a53]>",
		"883d4: CityRecord<ID=[5006848] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Reading] POP=(1056) LAT=(41.8394900000) LON=(-84.7480100000) S2=[883d77a0a78666d5]>",
		"883d4: CityRecord<ID=[5010780] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Spring Arbor] POP=(2881) LAT=(42.2050400000) LON=(-84.5527400000) S2=[883d393d44bc6eb5]>",
		"883d4: CityRecord<ID=[5010899] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Springport] POP=(790) LAT=(42.3783700000) LON=(-84.6985900000) S2=[883d4c706a00b23b]>",
		"883d4: CityRecord<ID=[5013156] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Vandercook Lake] POP=(4721) LAT=(42.1933700000) LON=(-84.3910700000) S2=[883d24615343ca47]>",
		"883d4: CityRecord<ID=[7259381] COUNTRY=[United States] PROVINCE-OR-STATE=[MI] CITY=[Manitou Beach-Devils Lake] POP=(2019) LAT=(41.9756500000) LON=(-84.2861600000) S2=[883d1d75f92a1b35]>",
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

	ci := NewTestCityIndex()

	err = ci.Load(gp, g, nil)
	log.PanicIf(err)

	// Do the query.

	clawsonCoordinates := []float64{42.53667, -83.15041}

	sourceName, visits, cr, err := ci.Nearest(clawsonCoordinates[0], clawsonCoordinates[1], true)
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

func TestCityIndex_getNearestPoint1(t *testing.T) {
	ci := NewTestCityIndex()

	originLatitude := 27.2974891
	originLongitude := -81.3871491

	queries := []VisitHistoryItem{
		VisitHistoryItem{
			Token: "aa",
			City: geoattractor.CityRecord{
				Latitude:  27.443239,
				Longitude: -81.429949,
			},
		},
		VisitHistoryItem{
			Token: "bb",
			City: geoattractor.CityRecord{
				Latitude:  27.038644,
				Longitude: -81.291909,
			},
		},
		VisitHistoryItem{
			Token: "cc",
			City: geoattractor.CityRecord{
				Latitude:  26.013582,
				Longitude: -80.542458,
			},
		},
	}

	vhi := ci.getNearestPoint(originLatitude, originLongitude, queries)
	if vhi.Token != "aa" {
		t.Fatalf("Result not correct: %v\n", vhi)
	}
}

func TestCityIndex_getNearestPoint2(t *testing.T) {
	ci := NewTestCityIndex()

	originLatitude := 26.00
	originLongitude := -80.50

	queries := []VisitHistoryItem{
		VisitHistoryItem{
			Token: "aa",
			City: geoattractor.CityRecord{
				Latitude:  27.443239,
				Longitude: -81.429949,
			},
		},
		VisitHistoryItem{
			Token: "bb",
			City: geoattractor.CityRecord{
				Latitude:  27.038644,
				Longitude: -81.291909,
			},
		},
		VisitHistoryItem{
			Token: "cc",
			City: geoattractor.CityRecord{
				Latitude:  26.013582,
				Longitude: -80.542458,
			},
		},
	}

	vhi := ci.getNearestPoint(originLatitude, originLongitude, queries)
	if vhi.Token != "cc" {
		t.Fatalf("Result not correct: %v\n", vhi)
	}
}

func TestCityIndex_getNearestPoint_OutOfOrder(t *testing.T) {
	ci := NewTestCityIndex()

	originLatitude := 27.2974891
	originLongitude := -81.3871491

	queries := []VisitHistoryItem{
		VisitHistoryItem{
			Token: "bb",
			City: geoattractor.CityRecord{
				Latitude:  27.038644,
				Longitude: -81.291909,
			},
		},
		VisitHistoryItem{
			Token: "cc",
			City: geoattractor.CityRecord{
				Latitude:  26.013582,
				Longitude: -80.542458,
			},
		},
		VisitHistoryItem{
			Token: "aa",
			City: geoattractor.CityRecord{
				Latitude:  27.443239,
				Longitude: -81.429949,
			},
		},
	}

	vhi := ci.getNearestPoint(originLatitude, originLongitude, queries)
	if vhi.Token != "aa" {
		t.Fatalf("Result not correct: %v\n", vhi)
	}
}
