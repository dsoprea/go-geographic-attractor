[![Build Status](https://travis-ci.org/dsoprea/go-geographic-attractor.svg?branch=master)](https://travis-ci.org/dsoprea/go-geographic-attractor)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-geographic-attractor/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-geographic-attractor?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-geographic-attractor?status.svg)](https://godoc.org/github.com/dsoprea/go-geographic-attractor/index)

# Overview

Produces the nearest major city to a given coordinate.


# Details

The purpose of this project is to provide an index that identifies the nearest urban center ("biggest city", "metropolitan area") very efficiently using Google's S2 algorithm as applied to the Earth. This basically subdivides the Earth into individual cells in a way that mitigates traditional distortion as well as optimizing how we interrelate smaller areas with larger areas.

This algorithm looks to within, approximately, thirty to forty miles of the given coordinates for a city with a population of at least 100,000. If one is found, that city is returned. If one is not found, the first city that *was* found is returned (so, not an urban center but the actual city that was searched). If no city was found, then error `geoattractorindex.ErrNoNearestCity` is returned.


# Algorithm Notes

Due to how the cells are calculated, this is only an approximation and the nearest city may sometimes be biased a little north/south/west/east of what you were expecting. However, the algorithm is very, very efficient and reduces a problem that is traditionally solved via clustering (very expensive) to a string-prefix search.

In other words, this algorithm is what you want if you can accept some minor approximation errors in exchange for instanteous searches rather than crunching numbers distributed across a cluster.


# Requirements

- A supported dataset. Currently, only [GeoNames](https://www.geonames.org) is supported. Browse to "Download" -> "[Free Gazetteer Data](http://download.geonames.org/export/dump)". Specifically, we require "countryInfo.txt" and "allCountries.zip" files.


# Usage

For usage examples, see the examples at [GoDoc](https://godoc.org/github.com/dsoprea/go-geographic-attractor).


# Tool

A command-line tool is also provided in order to test the index. This will load the index and then perform the search. As the index exists in memory, this is done at the top of every execution.


## Install

```
$ go get -t github.com/dsoprea/go-geographic-attractor/command/find_nearest_city
$ cd $GOPATH/src/github.com/dsoprea/go-geographic-attractor/command/find_nearest_city
$ go install
```


## Usage Examples

A simple query:

```
$ $GOPATH/bin/find_nearest_city --latitude 25.648315 --longitude -80.314120 --country-data-filepath countryInfo.txt --city-data-filepath allCountries.zip
Source: GeoNames
ID: 7170183
Country: United States
City: City of Hialeah
Population: 224669
Latitude: 25.8696300000
Longitude: -80.3045600000
```

Note that the tool is obviously just for testing as loading the index is [necessarily] expensive:

```
$ time $GOPATH/bin/find_nearest_city --latitude 25.648315 --longitude -80.314120 --country-data-filepath countryInfo.txt --city-data-filepath allCountries.zip
Source: GeoNames
ID: 7170183
Country: United States
City: City of Hialeah
Population: 224669
Latitude: 25.8696300000
Longitude: -80.3045600000

real	0m23.927s
user	0m28.365s
sys	0m0.346s
```

Print with increased verbosity. Specifically, this will print the concentric cells that are checked as we move from the smallest cell (with longer S2 cell IDs representing smaller, specific cells containing the nearest city to the given coordinates) outwards to larger cells (with smaller S2 cell IDs representing larger areas):

```
$ $GOPATH/bin/find_nearest_city --latitude 25.648315 --longitude -80.314120 --country-data-filepath countryInfo.txt --city-data-filepath allCountries.zip --verbose
VISIT( 0): 88d9c65: CityRecord<ID=[7172726] COUNTRY=[United States] CITY=[Village of Pinecrest] POP=(18223) LAT=(25.6650300000) LON=(-80.3042300000)>
VISIT( 1): 88d9c64: CityRecord<ID=[7172726] COUNTRY=[United States] CITY=[Village of Pinecrest] POP=(18223) LAT=(25.6650300000) LON=(-80.3042300000)>
VISIT( 2): 88d9c7: CityRecord<ID=[7172726] COUNTRY=[United States] CITY=[Village of Pinecrest] POP=(18223) LAT=(25.6650300000) LON=(-80.3042300000)>
VISIT( 3): 88d9c4: CityRecord<ID=[7171588] COUNTRY=[United States] CITY=[Town of Cutler Bay] POP=(40286) LAT=(25.5764800000) LON=(-80.3356600000)>
VISIT( 4): 88d9d: CityRecord<ID=[7317991] COUNTRY=[United States] CITY=[City of Homestead] POP=(60512) LAT=(25.4664000000) LON=(-80.4472300000)>
VISIT( 5): 88d9c: CityRecord<ID=[7170183] COUNTRY=[United States] CITY=[City of Hialeah] POP=(224669) LAT=(25.8696300000) LON=(-80.3045600000)>

Source: GeoNames
ID: 7170183
Country: United States
City: City of Hialeah
Population: 224669
Latitude: 25.8696300000
Longitude: -80.3045600000
```


Print as JSON:

```
$ $GOPATH/bin/find_nearest_city --latitude 25.648315 --longitude -80.314120 --country-data-filepath countryInfo.txt --city-data-filepath allCountries.zip --json
{
  "Result": {
    "id": "7170183",
    "country": "United States",
    "city": "City of Hialeah",
    "population": 224669,
    "latitude": 25.86963,
    "longitude": -80.30456
  },
  "Stats": {
    "unfiltered_records_parsed": 11857354,
    "records_added_to_index": 2061072,
    "records_updated_in_index": 60040
  }
}
```


No nearest city:

```
$ $GOPATH/bin/find_nearest_city --latitude 11.827416 --longitude -110.548982 --country-data-filepath countryInfo.txt --city-data-filepath allCountries.zip
No nearest city found.

$ echo $?
10
```
