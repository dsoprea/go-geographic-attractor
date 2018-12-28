package geoattractor

import (
	"path"
)

var (
	testAssetsPath string
)

func init() {
	testAssetsPath = path.Join(appPath, "test", "asset")
}
