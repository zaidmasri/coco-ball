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
	staticFileServer := http.FileServer(http.FS(static.FS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Routes
	// Auth
	mux.HandleFunc("GET /login", app.GetLogin())
	mux.HandleFunc("POST /login", app.PostLogin())
	mux.HandleFunc("GET /signup", app.GetSignup())
	mux.HandleFunc("POST /signup", app.PostSignup())
	mux.HandleFunc("GET /logout", app.GetLogout())

	// Profile
	mux.HandleFunc("GET /profile", app.GetProfile())

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

	// Payroll
	mux.HandleFunc("GET /plan/{id}/payroll", app.GetPayroll())

	// Sales Forecast
	mux.HandleFunc("GET /plan/{id}/sales-forecast", app.GetSalesForecast())

	// Operating Expenses
	mux.HandleFunc("GET /plan/{id}/operating-expenses", app.GetOpExpenses())

	// Cash Flow
	mux.HandleFunc("GET /plan/{id}/cash-flow", app.GetCashFlow())

	// Income Statement
	mux.HandleFunc("GET /plan/{id}/income-statement", app.GetIncomeStatement())

	// Balance Sheet
	mux.HandleFunc("GET /plan/{id}/balance-sheet", app.GetBalanceSheet())

	// Analytics
	mux.HandleFunc("GET /plan/{id}/analytics", app.GetAnalytics())

	// Catch-all fallback route for any unmatched URLs
	mux.HandleFunc("/", app.NotFound())

	// Wrap mux with middlewares
	var httpHandler http.Handler = mux
	httpHandler = app.AuthMiddleware(httpHandler)
	httpHandler = handlers.Logger(httpHandler)

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

	pages, err := fs.Glob(templates.FS, "pages/*.html")
	if err != nil {
		log.Fatal("error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		// Standalone pages that don't use base layout
		if name == "index.html" || name == "login.html" || name == "signup.html" || name == "profile.html" {
			ts, parseErr = template.ParseFS(templates.FS, page)
		} else {
			ts, parseErr = template.ParseFS(templates.FS, "base.html", "components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}

	return templateCache
}
