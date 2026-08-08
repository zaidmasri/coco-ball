// Package handlers has the http endpoints the application uses
package handlers

import (
	"net/http"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
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

	// Payroll - summary page, requires viewer access
	mux.Handle("GET /plan/{id}/payroll", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetPayrollSummary()(w, r)
	})))

	// Payroll - section landing/description page. "intro" comes before
	// {section} for the same reason as Starting Point's equivalent route
	// (see the comment above it): avoids a genuine ambiguity with
	// .../payroll-tax-rates/{step} that net/http's mux rejects at startup.
	mux.Handle("GET /plan/{id}/payroll/intro/{section}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetPayrollSectionIntro()(w, r)
	})))

	// Payroll: Salary Roles wizard
	mux.Handle("GET /plan/{id}/payroll/salary-roles", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalaryRoleList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalaryRoleNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalaryRoleFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/payroll/salary-roles/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalaryRoleStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalaryRoleStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalaryRoleDelete()(w, r)
	})))

	// Payroll: Benefits wizard
	mux.Handle("GET /plan/{id}/payroll/benefits", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetBenefitList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/benefits/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostBenefitNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/benefits/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostBenefitFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/payroll/benefits/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetBenefitStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/benefits/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostBenefitStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/benefits/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostBenefitDelete()(w, r)
	})))

	// Payroll: Payroll Tax Rates wizard (singleton per plan, no item ID)
	mux.Handle("GET /plan/{id}/payroll/payroll-tax-rates", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetPayrollTaxRatesEntry()(w, r)
	})))
	mux.Handle("GET /plan/{id}/payroll/payroll-tax-rates/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetPayrollTaxRatesStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/payroll/payroll-tax-rates/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostPayrollTaxRatesStep()(w, r)
	})))

	// Sales Forecast - summary page, requires viewer access
	mux.Handle("GET /plan/{id}/sales-forecast", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalesForecastSummary()(w, r)
	})))

	// Sales Forecast - section landing/description page (see Payroll's
	// equivalent route above for why "intro" is placed before {section}).
	mux.Handle("GET /plan/{id}/sales-forecast/intro/{section}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalesForecastSectionIntro()(w, r)
	})))

	// Sales Forecast: Products wizard
	mux.Handle("GET /plan/{id}/sales-forecast/products", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetProductList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast/products/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostProductNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast/products/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostProductFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/sales-forecast/products/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetProductStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast/products/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostProductStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast/products/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostProductDelete()(w, r)
	})))

	// Sales Forecast: Sales Growth Curve wizard (singleton per plan, no item ID)
	mux.Handle("GET /plan/{id}/sales-forecast/sales-growth-curve", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalesGrowthCurveEntry()(w, r)
	})))
	mux.Handle("GET /plan/{id}/sales-forecast/sales-growth-curve/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetSalesGrowthCurveStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/sales-forecast/sales-growth-curve/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostSalesGrowthCurveStep()(w, r)
	})))

	// Operating Expenses wizard - a single repeatable section, so
	// /plan/{id}/operating-expenses is directly the list page (no
	// multi-section summary/intro layer, unlike Payroll/Sales Forecast/
	// Cash Flow).
	mux.Handle("GET /plan/{id}/operating-expenses", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetOperatingExpenseList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/operating-expenses/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostOperatingExpenseNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/operating-expenses/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostOperatingExpenseFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/operating-expenses/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetOperatingExpenseStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/operating-expenses/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostOperatingExpenseStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/operating-expenses/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostOperatingExpenseDelete()(w, r)
	})))

	// Cash Flow - summary page, requires viewer access
	mux.Handle("GET /plan/{id}/cash-flow", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetCashFlowSummary()(w, r)
	})))

	// Cash Flow - section landing/description page (see Payroll's
	// equivalent route above for why "intro" is placed before {section}).
	mux.Handle("GET /plan/{id}/cash-flow/intro/{section}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetCashFlowSectionIntro()(w, r)
	})))

	// Cash Flow: Inventory Purchases wizard
	mux.Handle("GET /plan/{id}/cash-flow/inventory-purchases", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetInventoryPurchaseList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/inventory-purchases/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostInventoryPurchaseNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/inventory-purchases/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostInventoryPurchaseFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/cash-flow/inventory-purchases/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetInventoryPurchaseStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/inventory-purchases/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostInventoryPurchaseStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/inventory-purchases/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostInventoryPurchaseDelete()(w, r)
	})))

	// Cash Flow: Distributions wizard
	mux.Handle("GET /plan/{id}/cash-flow/distributions", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetDistributionList()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/distributions/new", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostDistributionNew()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/distributions/finish", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostDistributionFinish()(w, r)
	})))
	mux.Handle("GET /plan/{id}/cash-flow/distributions/{itemID}/{step}", app.RequireAccess(domain.Viewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.GetDistributionStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/distributions/{itemID}/{step}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostDistributionStep()(w, r)
	})))
	mux.Handle("POST /plan/{id}/cash-flow/distributions/{itemID}", app.RequireAccess(domain.Editor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.PostDistributionDelete()(w, r)
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
