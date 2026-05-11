// Package templates handles html rendering for web project
package templates

import "embed"

//go:embed *.html components/*.html pages/*.html
var FS embed.FS
