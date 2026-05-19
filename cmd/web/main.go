package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/handlers"
	"github.com/zaidmasri/business-planning-tool/internal/static"
	"github.com/zaidmasri/business-planning-tool/internal/store"
	"github.com/zaidmasri/business-planning-tool/internal/templates"
)

var templateCache map[string]*template.Template

func init() {
	templateCache = make(map[string]*template.Template)

	pages, err := fs.Glob(templates.FS, "pages/*.html")
	if err != nil {
		log.Fatal("error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		if name == "index.html" {
			ts, parseErr = template.ParseFS(templates.FS, page)
		} else {
			ts, parseErr = template.ParseFS(templates.FS, "base.html", "components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}
}

func main() {
	// Initialize the store
	memoryStore := store.NewMemoryStore()

	// Initialize the handlers application struct with our dependencies
	app := handlers.NewApp(memoryStore, templateCache)

	mux := http.NewServeMux()

	// Static files
	staticFileServer := http.FileServer(http.FS(static.FS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Routes
	// Root Page
	mux.HandleFunc("GET /{$}", app.GetRoot())

	// New plan
	mux.HandleFunc("POST /plan/setup", app.PostSetup())

	// Setup Page
	mux.HandleFunc("GET /plan/{id}/setup", app.GetSetup())
	mux.HandleFunc("POST /plan/{id}/setup", app.PostUpdateSetup())

	// Starting Point
	mux.HandleFunc("GET /plan/{id}/starting-point", app.GetStartingPoint())
	mux.HandleFunc("POST /plan/{id}/starting-point", app.PostStartingPoint())
	// Catch-all fallback route for any unmatched URLs
	mux.HandleFunc("/", app.NotFound())
	// Configure a robust http server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handlers.Logger(mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
