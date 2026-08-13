package tools

import (
	"github.com/ognerezov/hot-core/console"
	"strings"
)

const (
	VersionPlaceholder = "unknown"
	DefaultPass        = "./VERSION"
)

var (
	AppVersion = VersionPlaceholder
)

func ReadVersion(path string) string {
	if AppVersion != VersionPlaceholder {
		return AppVersion
	}
	bytes, err := ReadFile(path)
	if err != nil {
		console.RedPrintln(err.Error())
		return AppVersion
	}
	AppVersion = strings.TrimSpace(string(bytes))
	return AppVersion
}
