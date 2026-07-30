// Package views handles html rendering for web project
package views

import "embed"

//go:embed templates/*.html templates/components/*.html templates/pages/*.html
var TemplatesFS embed.FS
