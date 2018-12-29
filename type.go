package geoattractor

import (
    "fmt"
)

type CityRecord struct {
    Id                  string
    Country             string
    City                string
    Population          uint64
    Latitude, Longitude float64
}

func (cr CityRecord) String() string {
    return fmt.Sprintf("CityRecord<ID=[%s] COUNTRY=[%s] CITY=[%s] POP=(%d) LAT=(%.10f) LON=(%.10f)>", cr.Id, cr.Country, cr.City, cr.Population, cr.Latitude, cr.Longitude)
}
