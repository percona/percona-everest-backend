// Package public embeds the FE app into the BE
package public

import "embed"

// Static stores the latest version of the everest FE app
//
//go:embed dist/*
var Static embed.FS
