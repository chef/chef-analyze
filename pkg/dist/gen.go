// +build dist

// this file will be ignored for builds, but included for dependencies, it will
// generate automatically a 'global.go' with a set of global variables

package dist

import (
	_ "github.com/afiune/godist"
)

//go:generate go run github.com/afiune/godist global.go dist
