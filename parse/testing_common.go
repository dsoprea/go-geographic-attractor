package geoattractorparse

import (
	"path"
)

var (
	testAssetsPath string
)

func init() {
	testAssetsPath = path.Join(packagePath, "test", "asset")
}
