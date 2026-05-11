package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/static"
	"github.com/zaidmasri/business-planning-tool/internal/templates"
)

var templateCache map[string]*template.Template

func init() {
	templateCache = make(map[string]*template.Template)

	pages, err := fs.Glob(templates.FS, "pages/*.html")
	if err != nil {
		log.Fatal("Error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		// Treat index.html as a completely separate, standalone page
		if name == "index.html" {
			ts, parseErr = template.ParseFS(templates.FS, page)
		} else {
			// All other pages get the base layout and components
			ts, parseErr = template.ParseFS(templates.FS, "base.html", "components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("Error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}
}

// Helper to render from the cache safely
func renderPage(w http.ResponseWriter, cacheKey string, templateName string, data interface{}) {
	ts, ok := templateCache[cacheKey]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	// Execute the specific template requested
	err := ts.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	// Execute the standalone "index.html" file
	renderPage(w, "index.html", "index.html", map[string]string{
		"Title": "Business Planning Tool",
		"Path":  r.URL.Path,
	})
}

func handlerSetup(w http.ResponseWriter, r *http.Request) {
	// Execute "base" which will pull in the content from setup.html
	renderPage(w, "setup.html", "base", map[string]string{
		"Title": "Setup | Business Planning Tool",
		"Path":  r.URL.Path,
	})
}

func handlerStartingPoint(w http.ResponseWriter, r *http.Request) {
	// Execute "base" which will pull in the content from starting-point.html
	renderPage(w, "starting-point.html", "base", map[string]string{
		"Title": "Starting Point | Business Planning Tool",
		"Path":  r.URL.Path,
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()

	// Static Files
	staticFileServer := http.FileServer(http.FS(static.FS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Routes
	mux.HandleFunc("GET /{$}", handlerRoot)
	mux.HandleFunc("GET /setup", handlerSetup)
	mux.HandleFunc("GET /starting-point", handlerStartingPoint)

	// 3. Configure a robust HTTP Server with timeouts
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
