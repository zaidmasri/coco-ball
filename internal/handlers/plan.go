// Package handlers contains http functions for mux
package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/store"
)

// App holds the dependencies for all handlers
type App struct {
	Store         store.PlanStore
	TemplateCache map[string]*template.Template
}

// NewApp is the constructor to inject dependencies from main.go
func NewApp(s store.PlanStore, tc map[string]*template.Template) *App {
	return &App{
		Store:         s,
		TemplateCache: tc,
	}
}

// renderPage is now a method on App so it can access the TemplateCache
func (app *App) renderPage(w http.ResponseWriter, cacheKey string, templateName string, data interface{}) {
	ts, ok := app.TemplateCache[cacheKey]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	err := ts.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- HANDLERS (Exported with Capital Letters) ---
func (app *App) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.renderPage(w, "index.html", "index.html", map[string]interface{}{
			"Title": "Business Planning Tool",
			"Path":  r.URL.Path,
		})
	}
}

// POST /plan/setup (Saves a brand new plan)
func (app *App) PostSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		companyName := r.PostForm.Get("companyName")
		startMonth, _ := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, _ := strconv.Atoi(r.PostForm.Get("startYear"))

		newID := uuid.New()
		plan, err := domain.NewPlan(newID, companyName, startMonth, startYear)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = app.Store.Save(plan)
		if err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plan/"+newID.String()+"/starting-point", http.StatusSeeOther)
	}
}

// GET /plan/{id}/setup (Loads an existing plan into the form)
func (app *App) GetSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		planID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid Plan ID", http.StatusBadRequest)
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			http.Error(w, "Plan not found", http.StatusNotFound)
			return
		}

		app.renderPage(w, "setup.html", "base", map[string]interface{}{
			"Title": "Edit Setup | Business Planning Tool",
			"Path":  r.URL.Path,
			"Plan":  plan,
		})
	}
}

// GET /plan/{id}/starting-point
func (app *App) GetStartingPoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		planID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid Plan ID", http.StatusBadRequest)
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			http.Error(w, "Plan not found", http.StatusNotFound)
			return
		}

		app.renderPage(w, "starting-point.html", "base", map[string]interface{}{
			"Title": "Starting Point | Business Planning Tool",
			"Path":  r.URL.Path,
			"Plan":  plan,
		})
	}
}

// Exported Middleware
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}
