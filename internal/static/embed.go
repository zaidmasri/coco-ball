// Package static is used for adding css files to the web applicatiion
package static

import "embed"

//go:embed css js icons
var FS embed.FS
