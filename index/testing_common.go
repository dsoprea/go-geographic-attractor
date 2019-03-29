package geoattractorindex

import (
    "path"
)

var (
    testAssetsPath string
)

func NewTestCityIndex() *CityIndex {
    return NewCityIndex(DefaultMinimumLevelForUrbanCenterAttraction, DefaultUrbanCenterMinimumPopulation)
}

func init() {
    testAssetsPath = path.Join(packagePath, "test", "asset")
}
