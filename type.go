package geoattractor

import (
    "fmt"
    "io"
    "strconv"

    "encoding/gob"

    "github.com/golang/geo/s2"
    "github.com/randomingenuity/go-utility/geographic"
)

type CityRecord struct {
    Id            string  `json:"id"`
    Country       string  `json:"country"`
    ProvinceState string  `json:"province_or_state"`
    City          string  `json:"city"`
    Population    uint64  `json:"population"`
    Latitude      float64 `json:"latitude"`
    Longitude     float64 `json:"longitude"`
    Cell          s2.CellID
}

func (cr CityRecord) String() string {
    s2Token := cr.S2Cell().ToToken()
    return fmt.Sprintf("CityRecord<ID=[%s] COUNTRY=[%s] PROVINCE-OR-STATE=[%s] CITY=[%s] POP=(%d) LAT=(%.10f) LON=(%.10f) S2=[%s]>", cr.Id, cr.Country, cr.ProvinceState, cr.City, cr.Population, cr.Latitude, cr.Longitude, s2Token)
}

func (cr CityRecord) CityAndProvinceState() string {
    name := cr.City

    // Only append ProvinceState if not [wholly] a number.

    _, err := strconv.Atoi(cr.ProvinceState)
    if err != nil {
        name += fmt.Sprintf(", %s", cr.ProvinceState)
    }

    return name
}

func (cr CityRecord) S2Cell() s2.CellID {
    if uint64(cr.Cell) == 0 {
        cr.Cell = rigeo.S2CellFromCoordinates(cr.Latitude, cr.Longitude)
    }

    return cr.Cell
}

type CityRecordCb func(cr CityRecord) (err error)

type CityRecordSource interface {
    Parse(r io.Reader, cb CityRecordCb) (recordsCount int, err error)
    Name() string
}

func init() {
    gob.Register(CityRecord{})
}
