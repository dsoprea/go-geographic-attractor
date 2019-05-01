package geoattractorparse

import (
	"fmt"
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
	log.PanicIf(err)

	defer f.Close()

	actual := make([]string, 0)
	cb := func(cr geoattractor.CityRecord) (err error) {
		actual = append(actual, cr.String())

		return nil
	}

	recordsCount, err := gp.Parse(f, cb)
	log.PanicIf(err)

	if recordsCount != 35 {
		t.Fatalf("Number of records read is not correct: (%d)", recordsCount)
	}

	expected := []string{
		"CityRecord<ID=[3038999] COUNTRY=[Andorra] PROVINCE-OR-STATE=[02] CITY=[Soldeu] POP=(602) LAT=(42.5768800000) LON=(1.6676900000) S2=[12af62d7c07ae2a3]>",
		"CityRecord<ID=[3039154] COUNTRY=[Andorra] PROVINCE-OR-STATE=[02] CITY=[El Tarter] POP=(1052) LAT=(42.5795200000) LON=(1.6536200000) S2=[12af631e48015365]>",
		"CityRecord<ID=[3039163] COUNTRY=[Andorra] PROVINCE-OR-STATE=[06] CITY=[Sant Julià de Lòria] POP=(8022) LAT=(42.4637200000) LON=(1.4912900000) S2=[12a5f482b74a8b6f]>",
		"CityRecord<ID=[3039604] COUNTRY=[Andorra] PROVINCE-OR-STATE=[03] CITY=[Pas de la Casa] POP=(2363) LAT=(42.5427700000) LON=(1.7336100000) S2=[12af7d8b589b0e07]>",
		"CityRecord<ID=[3039678] COUNTRY=[Andorra] PROVINCE-OR-STATE=[05] CITY=[Ordino] POP=(3066) LAT=(42.5562300000) LON=(1.5331900000) S2=[12af6013cc4a202f]>",
		"CityRecord<ID=[3040051] COUNTRY=[Andorra] PROVINCE-OR-STATE=[08] CITY=[les Escaldes] POP=(15853) LAT=(42.5072900000) LON=(1.5341400000) S2=[12a58adbc21b2ba3]>",
		"CityRecord<ID=[3040132] COUNTRY=[Andorra] PROVINCE-OR-STATE=[04] CITY=[la Massana] POP=(7211) LAT=(42.5449900000) LON=(1.5148300000) S2=[12af5ff942041fdf]>",
		"CityRecord<ID=[3040140] COUNTRY=[Andorra] PROVINCE-OR-STATE=[02] CITY=[l'Aldosa de canillo] POP=(195) LAT=(42.5789500000) LON=(1.6290200000) S2=[12af63ad36ab439b]>",
		"CityRecord<ID=[3040141] COUNTRY=[Andorra] PROVINCE-OR-STATE=[04] CITY=[l'Aldosa] POP=(594) LAT=(42.5439100000) LON=(1.5228900000) S2=[12af6000fb372317]>",
		"CityRecord<ID=[3040686] COUNTRY=[Andorra] PROVINCE-OR-STATE=[03] CITY=[Encamp] POP=(11223) LAT=(42.5347400000) LON=(1.5801400000) S2=[12a58a0ca28aa2f1]>",
		"CityRecord<ID=[3041204] COUNTRY=[Andorra] PROVINCE-OR-STATE=[02] CITY=[Canillo] POP=(3292) LAT=(42.5676000000) LON=(1.5975600000) S2=[12af61883bca727d]>",
		"CityRecord<ID=[3041519] COUNTRY=[Andorra] PROVINCE-OR-STATE=[04] CITY=[Arinsal] POP=(1419) LAT=(42.5720500000) LON=(1.4845300000) S2=[12af5fa85ff003e3]>",
		"CityRecord<ID=[3041563] COUNTRY=[Andorra] PROVINCE-OR-STATE=[07] CITY=[Andorra la Vella] POP=(20430) LAT=(42.5077900000) LON=(1.5210900000) S2=[12a58ad34a0ad3ab]>",
		"CityRecord<ID=[7302102] COUNTRY=[Andorra] PROVINCE-OR-STATE=[07] CITY=[La Margineda] POP=(155) LAT=(42.4859900000) LON=(1.4902400000) S2=[12a5f4f24722170d]>",
		"CityRecord<ID=[10630523] COUNTRY=[Andorra] PROVINCE-OR-STATE=[04] CITY=[Puiol del Piu] POP=(400) LAT=(42.5652000000) LON=(1.4915900000) S2=[12af5fa3520b6311]>",
		"CityRecord<ID=[290594] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[07] CITY=[Umm al Qaywayn] POP=(44411) LAT=(25.5647300000) LON=(55.5551700000) S2=[3ef5fe5ef632c5e3]>",
		"CityRecord<ID=[291074] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[05] CITY=[Ras al-Khaimah] POP=(115949) LAT=(25.7895300000) LON=(55.9432000000) S2=[3ef6715330c8e03f]>",
		"CityRecord<ID=[291279] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[01] CITY=[Muzayri‘] POP=(10000) LAT=(23.1435500000) LON=(53.7881000000) S2=[3e66df70821ee015]>",
		"CityRecord<ID=[291339] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[06] CITY=[Murbaḩ] POP=(2000) LAT=(25.2762300000) LON=(56.3625600000) S2=[3ef4506055c17047]>",
		"CityRecord<ID=[291696] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[06] CITY=[Khawr Fakkān] POP=(33575) LAT=(25.3313200000) LON=(56.3419900000) S2=[3ef44fb5f49b1f4d]>",
		"CityRecord<ID=[292223] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[03] CITY=[Dubai] POP=(1137347) LAT=(25.0657000000) LON=(55.1712800000) S2=[3e5f6c8d52f684a7]>",
		"CityRecord<ID=[292231] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[04] CITY=[Dibba Al-Fujairah] POP=(30000) LAT=(25.5924600000) LON=(56.2617600000) S2=[3ef43245709559c9]>",
		"CityRecord<ID=[292239] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[04] CITY=[Dibba Al-Hisn] POP=(26395) LAT=(25.6195500000) LON=(56.2729100000) S2=[3ef4320c8c591929]>",
		"CityRecord<ID=[292672] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[06] CITY=[Sharjah] POP=(543733) LAT=(25.3373700000) LON=(55.4120600000) S2=[3e5f596791d602f3]>",
		"CityRecord<ID=[292688] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[01] CITY=[Ar Ruways] POP=(16000) LAT=(24.1102800000) LON=(52.7305600000) S2=[3e5d2b1394745fe7]>",
		"CityRecord<ID=[292878] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[04] CITY=[Al Fujayrah] POP=(62415) LAT=(25.1164100000) LON=(56.3414100000) S2=[3ef4f85930eb4671]>",
		"CityRecord<ID=[292913] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[01] CITY=[Al Ain] POP=(408733) LAT=(24.1916700000) LON=(55.7605600000) S2=[3e8ab73288163e53]>",
		"CityRecord<ID=[292932] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[02] CITY=[Ajman] POP=(226172) LAT=(25.4111100000) LON=(55.4350400000) S2=[3e5f5764bc00f04f]>",
		"CityRecord<ID=[292953] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[06] CITY=[Adh Dhayd] POP=(24716) LAT=(25.2881200000) LON=(55.8815700000) S2=[3ef5be65958120b3]>",
		"CityRecord<ID=[292968] COUNTRY=[United Arab Emirates] PROVINCE-OR-STATE=[01] CITY=[Abu Dhabi] POP=(603492) LAT=(24.4666700000) LON=(54.3666700000) S2=[3e5e66184786f8cf]>",
		"CityRecord<ID=[1120483] COUNTRY=[Afghanistan] PROVINCE-OR-STATE=[11] CITY=[Kuhsān] POP=(12087) LAT=(34.6538900000) LON=(61.1977800000) S2=[3f1623ae3b5e8097]>",
		"CityRecord<ID=[1120487] COUNTRY=[Afghanistan] PROVINCE-OR-STATE=[33] CITY=[Tukzār] POP=(12021) LAT=(35.9483100000) LON=(66.4213200000) S2=[3f3397869e6e5d63]>",
		"CityRecord<ID=[1120711] COUNTRY=[Afghanistan] PROVINCE-OR-STATE=[11] CITY=[Zindah Jān] POP=(10104) LAT=(34.3426400000) LON=(61.7467500000) S2=[3f3d1287f1a2e9bd]>",
		"CityRecord<ID=[1120863] COUNTRY=[Afghanistan] PROVINCE-OR-STATE=[29] CITY=[Zarghūn Shahr] POP=(13737) LAT=(32.8473400000) LON=(68.4457300000) S2=[38d593f140012be1]>",
		"CityRecord<ID=[1120879] COUNTRY=[Afghanistan] PROVINCE-OR-STATE=[29] CITY=[Zaṟah Sharan] POP=(7366) LAT=(33.1464100000) LON=(68.7921300000) S2=[38d5da394e4c5835]>",
	}

	if reflect.DeepEqual(actual, expected) == false {
		for _, visit := range actual {
			fmt.Printf("%s\n", visit)
		}

		t.Fatalf("Results not expected.")
	}
}
