package geoattractor

import (
    "fmt"
    "io"
)

type CityRecord struct {
    Id            string  `json:"id"`
    Country       string  `json:"country"`
    ProvinceState string  `json:"province_or_state"`
    City          string  `json:"city"`
    Population    uint64  `json:"population"`
    Latitude      float64 `json:"latitude"`
    Longitude     float64 `json:"longitude"`
}

func (cr CityRecord) String() string {
    return fmt.Sprintf("CityRecord<ID=[%s] COUNTRY=[%s] PROVINCE-OR-STATE=[%s] CITY=[%s] POP=(%d) LAT=(%.10f) LON=(%.10f)>", cr.Id, cr.Country, cr.ProvinceState, cr.City, cr.Population, cr.Latitude, cr.Longitude)
}

type CityRecordCb func(cr CityRecord) (err error)

type CityRecordSource interface {
    Parse(r io.Reader, cb CityRecordCb) (recordsCount int, err error)
    Name() string
}
