package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var value string

func String() string {
	return strings.TrimSpace(value)
}
