package main

import (
	_ "embed"
	"strings"
)

const defaultVersion = "0.0.0"

//go:embed VERSION.txt
var embeddedVersion string

func appVersion() string {
	version := strings.TrimSpace(embeddedVersion)
	if version == "" {
		return defaultVersion
	}
	return version
}
