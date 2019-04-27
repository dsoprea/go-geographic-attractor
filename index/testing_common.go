package geoattractorindex

import (
    "path"

    "io/ioutil"

    "github.com/dsoprea/go-logging"
)

var (
    testAssetsPath string
)

func NewTestCityIndex() (*CityIndex, string) {
    f, err := ioutil.TempFile("", "TestKvPut*")
    log.PanicIf(err)

    filepath := f.Name()
    ci := NewCityIndex(filepath, DefaultMinimumLevelForUrbanCenterAttraction, DefaultUrbanCenterMinimumPopulation)

    return ci, filepath
}

func init() {
    testAssetsPath = path.Join(packagePath, "test", "asset")
}
