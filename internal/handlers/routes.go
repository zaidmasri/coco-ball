// Package handlers has the http endpoints the application uses
package handlers

import (
	"net/http"

	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// RegisterRoutes registers all application routes on the provided mux
func (app *App) RegisterRoutes(mux *http.ServeMux) {
	// Auth
	mux.HandleFunc("GET /login", app.GetLogin())
	mux.HandleFunc("POST /login", app.PostLogin())
	mux.HandleFunc("GET /signup", app.GetSignup())
	mux.HandleFunc("POST /signup", app.PostSignup())
	mux.HandleFunc("POST /logout", app.PostLogout())

	// Profile
	mux.HandleFunc("GET /profile", app.GetProfile())

	// Root Page
	mux.HandleFunc("GET /{$}", app.GetRoot())

	// New plan
	mux.HandleFunc("POST /plan/setup", app.PostSetup())

	// Setup Page - requires viewer access
	mux.Handle("GET /plan/{id}/setup", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSetup()(w, r)
	})))
	mux.Handle("POST /plan/{id}/setup", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostUpdateSetup()(w, r)
	})))

	// Starting Point - summary page, requires viewer access
	mux.Handle("GET /plan/{id}/starting-point", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetStartingPointSummary()(w, r)
	})))

	// Starting Point - section landing/description page, shown from the
	// summary's "Get Started" link before a not-yet-started section.
	// "intro" comes before {section} (not after) so this can't collide
	// with /starting-point/cash-on-hand/{step} - a wildcard-then-literal
	// pattern here would tie with that literal-then-wildcard one on paths
	// like ".../cash-on-hand/intro", which net/http's mux rejects at
	// startup as genuinely ambiguous.
	mux.Handle("GET /plan/{id}/starting-point/intro/{section}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetStartingPointSectionIntro()(w, r)
	})))

	// Starting Point: Fixed Assets wizard
	mux.Handle("GET /plan/{id}/starting-point/fixed-assets", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetFixedAssetList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFixedAssetNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFixedAssetFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/starting-point/fixed-assets/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetFixedAssetStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFixedAssetStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFixedAssetDelete()(w, r)
	})))

	// Starting Point: Startup Costs wizard
	mux.Handle("GET /plan/{id}/starting-point/startup-costs", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetStartupCostList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostStartupCostNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostStartupCostFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/starting-point/startup-costs/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetStartupCostStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostStartupCostStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostStartupCostDelete()(w, r)
	})))

	// Starting Point: Funding Sources wizard
	mux.Handle("GET /plan/{id}/starting-point/funding-sources", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetFundingSourceList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFundingSourceNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFundingSourceFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/starting-point/funding-sources/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetFundingSourceStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFundingSourceStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostFundingSourceDelete()(w, r)
	})))

	// Starting Point: Cash on Hand wizard (singleton per plan, no item ID)
	mux.Handle("GET /plan/{id}/starting-point/cash-on-hand", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetCashOnHandEntry()(w, r)
	})))
	mux.Handle("GET /plan/{id}/starting-point/cash-on-hand/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetCashOnHandStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/starting-point/cash-on-hand/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostCashOnHandStep()(w, r)
	})))

	// Invites - only the plan owner can invite collaborators
	mux.Handle("POST /plan/{id}/invites", app.RequireAccess(domain.Owner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostCreateInvite()(w, r)
	})))
	// Accept/reject are keyed by invite ID, not plan ID, so auth is checked
	// inside the handler rather than via RequireAccess.
	mux.HandleFunc("POST /invites/{id}/accept", app.PostAcceptInvite())
	mux.HandleFunc("POST /invites/{id}/reject", app.PostRejectInvite())

	// Payroll - requires viewer access for GET, editor for POST
	mux.Handle("GET /plan/{id}/payroll", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetPayroll()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostPayroll()(w, r)
	})))

	// Sales Forecast - requires viewer access for GET, editor for POST
	mux.Handle("GET /plan/{id}/sales-forecast", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalesForecast()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalesForecast()(w, r)
	})))

	// Operating Expenses - requires viewer access for GET, editor for POST
	mux.Handle("GET /plan/{id}/operating-expenses", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetOpExpenses()(w, r)
	})))
	mux.Handle("POST /plan/{id}/operating-expenses", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostOpExpenses()(w, r)
	})))

	// Cash Flow - requires viewer access for GET, editor for POST
	mux.Handle("GET /plan/{id}/cash-flow", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetCashFlow()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostCashFlow()(w, r)
	})))

	// Income Statement - requires viewer access
	mux.Handle("GET /plan/{id}/income-statement", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetIncomeStatement()(w, r)
	})))

	// Balance Sheet - requires viewer access
	mux.Handle("GET /plan/{id}/balance-sheet", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetBalanceSheet()(w, r)
	})))

	// Analytics - requires viewer access
	mux.Handle("GET /plan/{id}/analytics", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetAnalytics()(w, r)
	})))

	// Delete plan - requires owner access
	mux.Handle("POST /plan/{id}/delete", app.RequireAccess(domain.Owner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostDeletePlan()(w, r)
	})))

	// Catch-all fallback route for any unmatched URLs
	mux.HandleFunc("/", app.NotFound())
}
