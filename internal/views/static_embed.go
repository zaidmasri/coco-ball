// Package views is used for adding css files to the web applicatiion
package views

import "embed"

//go:embed static/css static/js static/icons
var StaticFS embed.FS
