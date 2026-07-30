package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/handlers"
	"github.com/zaidmasri/business-planning-tool/internal/middleware"
	"github.com/zaidmasri/business-planning-tool/internal/store"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

func serve(dbPath, port string) {
	// Initialize the database store
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer sqliteStore.Close()

	// Load templates
	templateCache := loadTemplates()

	// Initialize the handlers application struct with our dependencies
	app := handlers.NewApp(sqliteStore, templateCache)

	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(views.StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static fs: %v", err)
	}
	staticFileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Register all application routes
	app.RegisterRoutes(mux)

	// Wrap mux with middlewares
	var httpHandler http.Handler = mux
	httpHandler = app.AuthMiddleware(httpHandler)
	httpHandler = middleware.Logger(httpHandler)

	// Configure a robust http server
	srv := &http.Server{
		Addr:         port,
		Handler:      httpHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on %s", port)
	log.Fatal(srv.ListenAndServe())
}

func loadTemplates() map[string]*template.Template {
	templateCache := make(map[string]*template.Template)

	pages, err := fs.Glob(views.TemplatesFS, "templates/pages/*.html")
	if err != nil {
		log.Fatal("error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		// Standalone pages that don't use base layout
		if name == "index.html" || name == "login.html" || name == "signup.html" || name == "profile.html" {
			ts, parseErr = template.ParseFS(views.TemplatesFS, page)
		} else {
			ts, parseErr = template.ParseFS(views.TemplatesFS, "templates/base.html", "templates/components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}

	return templateCache
}
