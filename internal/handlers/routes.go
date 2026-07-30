// Package handlers has the http endpoints the application uses
package handlers

import "net/http"

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
}
