// Package views handles html rendering for web project
package views

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
)

//go:embed templates/*.html templates/components/*.html templates/pages/*.html
var TemplatesFS embed.FS

// standalonePages don't use the shared base layout and are parsed on their own.
var standalonePages = map[string]bool{
	"index.html":   true,
	"login.html":   true,
	"signup.html":  true,
	"profile.html": true,
}

// LoadTemplates parses every page template under templates/pages into a cache
// keyed by base filename, pairing each with base.html and the shared
// components unless it's a standalone page. Callers (the production server
// and tests that need real rendered output) share this so template parsing
// only happens in one place.
func LoadTemplates() map[string]*template.Template {
	templateCache := make(map[string]*template.Template)

	pages, err := fs.Glob(TemplatesFS, "templates/pages/*.html")
	if err != nil {
		log.Fatal("error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		if standalonePages[name] {
			ts, parseErr = template.New(name).Funcs(TemplateFuncs).ParseFS(TemplatesFS, page)
		} else {
			ts, parseErr = template.New("base.html").Funcs(TemplateFuncs).ParseFS(TemplatesFS, "templates/base.html", "templates/components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}

	return templateCache
}
