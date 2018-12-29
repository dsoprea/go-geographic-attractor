package geoattractorindex

import (
    "os"
    "path"
)

var (
    appPath     string
    packagePath string
)

func init() {
    goPath := os.Getenv("GOPATH")
    appPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-geographic-attractor")
    packagePath = path.Join(appPath, "index")
}
