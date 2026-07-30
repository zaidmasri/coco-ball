// Package handlers contains http functions for mux
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/store"
	"github.com/zaidmasri/business-planning-tool/internal/views"
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

// renderErrorPage is a helper to render error pages
func (app *App) renderErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	page := views.BuildErrorPage(r, statusCode, message)
	views.RenderErrorPageWithStatus(w, app.TemplateCache, page, statusCode)
}

// NotFound is user for GET requests (Custom 404 Catch-All for bad URLs)
func (app *App) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.renderErrorPage(w, r, http.StatusNotFound, "The URL you entered does not exist. Please check the address and try again.")
	}
}

func (app *App) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			// Unauthenticated access
			page := views.IndexPage{
				BasePage: views.BasePage{
					Title: "Business Planning Tool",
					Path:  r.URL.Path,
					User:  nil,
				},
				Plans: []*domain.Plan{},
			}
			views.RenderIndexPage(w, app.TemplateCache, page)
			return
		}

		// Fetch plans the user has access to
		plans, err := app.Store.GetUserPlans(user.ID())
		if err != nil {
			plans = []*domain.Plan{}
		}

		page := views.BuildIndexPage(r, user, plans)
		views.RenderIndexPage(w, app.TemplateCache, page)
	}
}

// PostSetup POST /plan/setup (Saves a brand new plan)
func (app *App) PostSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		companyName := r.PostForm.Get("companyName")
		startMonth, _ := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, _ := strconv.Atoi(r.PostForm.Get("startYear"))

		newID := uuid.New()
		plan, err := domain.NewPlan(newID, companyName, startMonth, startYear, user.ID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = app.Store.Save(plan)
		if err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		// Grant owner access to the creator
		if err := app.Store.GrantAccess(newID, user.ID(), domain.Owner); err != nil {
			log.Printf("Failed to grant owner access: %v", err)
			http.Error(w, "Failed to setup plan access", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plan/"+newID.String()+"/starting-point", http.StatusSeeOther)
	}
}

// PostUpdateSetup POST /plan/{id}/setup (Updates an existing plan)
func (app *App) PostUpdateSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		// Group extraction and type conversion together
		companyName := r.PostForm.Get("companyName")
		startMonth, errMonth := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, errYear := strconv.Atoi(r.PostForm.Get("startYear"))

		// Combine the error checks for related fields
		if errMonth != nil || errYear != nil {
			http.Error(w, "Invalid month or year provided", http.StatusBadRequest)
			return
		}

		// Use the error message directly from your domain logic
		if err := plan.ChangeCoreDetails(companyName, startMonth, startYear); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.Store.Save(plan); err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		// Redirect to the exact same URL, forcing a GET request
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

// GetSetup GET /plan/{id}/setup (Loads an existing plan into the form)
func (app *App) GetSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
			return
		}

		page := views.BuildSetupPage(r, user, plan)
		views.RenderSetupPage(w, app.TemplateCache, page)
	}
}

// GetStartingPoint GET /plan/{id}/starting-point
func (app *App) GetStartingPoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
			return
		}

		page := views.BuildStartingPointPage(r, user, plan)
		views.RenderStartingPointPage(w, app.TemplateCache, page)
	}
}

// PostStartingPoint POST /plan/{id}/starting-point
func (app *App) PostStartingPoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// System-level errors (Malformed requests)
		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm Error: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		user := GetUserFromContext(r)

		// --- USER-FACING VALIDATION LOGIC ---

		// Helper function: Renders the starting-point page with an error message
		renderError := func(errMsg string, statusCode int) {
			w.WriteHeader(statusCode)
			page := views.BuildStartingPointPageWithError(r, user, plan, errMsg)
			views.RenderStartingPointPage(w, app.TemplateCache, page)
		}

		// 1. Clear existing data so we cleanly overwrite it
		plan.ClearStartingPoint()

		// 2. Parse Fixed Assets
		faTypes := r.PostForm["fa_type[]"]
		faAmounts := r.PostForm["fa_amount[]"]
		faDeprs := r.PostForm["fa_depreciation[]"]
		faMethods := r.PostForm["fa_depr_method[]"] // Grab the new dropdown data

		for i, assetName := range faTypes {
			// Protect against mismatched DOM arrays
			if i >= len(faAmounts) || i >= len(faDeprs) || i >= len(faMethods) {
				break
			}

			amt, _ := strconv.ParseFloat(faAmounts[i], 64)
			deprYears, _ := strconv.Atoi(faDeprs[i])
			methodStr := faMethods[i]

			// Map the string from the HTML select to your Domain type
			var deprMethod domain.DepreciationMethod
			if methodStr == string(domain.DoubleDeclining) {
				deprMethod = domain.DoubleDeclining
			} else if methodStr == string(domain.None) {
				deprMethod = domain.None
				deprYears = 0 // Land doesn't depreciate
			} else {
				deprMethod = domain.StraightLine
			}

			// Strategy 2: Smart Default (If they left it blank and it IS depreciable)
			if deprMethod != domain.None && deprYears <= 0 {
				deprYears = 5
			}

			if amt > 0 && strings.TrimSpace(assetName) != "" {
				err := plan.AddCapitalPurchase(domain.CapitalAsset{
					Name:               assetName,
					PurchaseCost:       domain.Money(amt),
					UsefulLifeMonths:   deprYears * 12, // Domain expects months
					DepreciationMethod: deprMethod,
				})
				// Strategy 4: Bubble up domain errors to the UI
				if err != nil {
					renderError(fmt.Sprintf("Error saving asset '%s': %v", assetName, err), http.StatusBadRequest)
					return
				}
			}
		}

		// 3. Parse Operating Capital
		ocTypes := r.PostForm["oc_type[]"]
		ocAmounts := r.PostForm["oc_amount[]"]

		for i, capName := range ocTypes {
			if i >= len(ocAmounts) {
				break
			}
			amt, _ := strconv.ParseFloat(ocAmounts[i], 64)
			if amt > 0 && strings.TrimSpace(capName) != "" {
				plan.AddStartupCost(capName, domain.Money(amt))
			}
		}

		// 4. Parse Funding Sources
		fundTypes := r.PostForm["fund_type[]"]
		fundAmounts := r.PostForm["fund_amount[]"]
		fundRates := r.PostForm["fund_rate[]"]
		fundTerms := r.PostForm["fund_term[]"]

		for i, fundName := range fundTypes {
			if i >= len(fundAmounts) || i >= len(fundRates) || i >= len(fundTerms) {
				break
			}
			amt, _ := strconv.ParseFloat(fundAmounts[i], 64)
			rate, _ := strconv.ParseFloat(fundRates[i], 64)
			term, _ := strconv.Atoi(fundTerms[i])

			if amt > 0 && strings.TrimSpace(fundName) != "" {
				plan.AddFundingSource(fundName, domain.Money(amt), rate/100.0, term)
			}
		}

		// 5. Parse Starting Balances (Static Fields)
		parseMoney := func(key string) domain.Money {
			val, _ := strconv.ParseFloat(r.PostForm.Get(key), 64)
			return domain.Money(val)
		}

		plan.SetStartingBalances(
			parseMoney("coh_cash_amount"),
			parseMoney("coh_ar_amount"),
			parseMoney("coh_pe_amount"),
			parseMoney("coh_ap_amount"),
			parseMoney("coh_ae_amount"),
		)

		// 6. Save the aggregate
		if err := app.Store.Save(plan); err != nil {
			log.Printf("Store Save Error: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		// Redirect on success
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func (app *App) GetPayroll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildPayrollPage(r, GetUserFromContext(r), plan)
		views.RenderPayrollPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetSalesForecast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildSalesForecastPage(r, GetUserFromContext(r), plan)
		views.RenderSalesForecastPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetOpExpenses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildOpExpensesPage(r, GetUserFromContext(r), plan)
		views.RenderOpExpensesPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetCashFlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildCashFlowPage(r, GetUserFromContext(r), plan)
		views.RenderCashFlowPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetIncomeStatement() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildIncomeStatementPage(r, GetUserFromContext(r), plan)
		views.RenderIncomeStatementPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetBalanceSheet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildBalanceSheetPage(r, GetUserFromContext(r), plan)
		views.RenderBalanceSheetPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildAnalyticsPage(r, GetUserFromContext(r), plan)
		views.RenderAnalyticsPage(w, app.TemplateCache, page)
	}
}

// parsePlanID extracts and parses the plan ID from the request path
func parsePlanID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue("id"))
}
