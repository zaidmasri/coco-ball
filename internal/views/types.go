package views

import (
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// BasePage contains common data for all pages
type BasePage struct {
	Title string
	Path  string
	User  *domain.User

	// StartingPointComplete/PayrollComplete/SalesForecastComplete/
	// OperatingExpensesComplete/CashFlowComplete each report whether every
	// sub-section of that hub is complete. The sidebar (rendered on every
	// page) uses these to decide whether each hub's nav icon shows filled
	// or outline.
	StartingPointComplete     bool
	PayrollComplete           bool
	SalesForecastComplete     bool
	OperatingExpensesComplete bool
	CashFlowComplete          bool
}

// IndexPage is the data for the home page
type IndexPage struct {
	BasePage
	Plans          []*domain.Plan
	PendingInvites []InviteSummary
}

// InviteSummary is a display-friendly view of a pending invite, used on the
// landing page onboarding section.
type InviteSummary struct {
	Invite      *domain.PlanInvite
	PlanName    string
	InviterName string
}

// SetupPage is the data for the plan setup page
type SetupPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
	Invites      []*domain.PlanInvite
	IsOwner      bool
}

// Starting Point/Payroll/Sales Forecast/Cash Flow's page types, builders,
// and renderers live in their own files (starting_point.go, payroll.go,
// sales_forecast.go, cash_flow.go), not here - they're numerous enough
// (summary, N repeatable sections x list/step/add-another, singleton
// sections, section intro) to warrant their own files rather than growing
// this one. Most of their page shapes are shared generic types defined in
// wizard_shared.go. Operating Expenses' page types, builders, and
// renderers live in operating_expenses.go, following the same pattern.

// IncomeStatementPage is the data for the income statement page
type IncomeStatementPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string

	Years     []domain.AnnualFinancials
	Products  []domain.ProductFinancials
	OpExLines []domain.OpExLine
}

// BalanceSheetPage is the data for the balance sheet page
type BalanceSheetPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string

	Years    []domain.BalanceSheetSnapshot
	Balances bool // whether Assets == Liabilities+Equity within rounding tolerance
}

// AnalyticsPage is the data for the analytics page
type AnalyticsPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string

	Breakeven domain.BreakevenAnalysis
	Ratios    []domain.FinancialRatios
	Years     []domain.BalanceSheetSnapshot
	Annuals   []domain.AnnualFinancials

	AssetDepreciation []domain.AssetDepreciation
	LoanAmortization  []domain.LoanYearSummary

	OwnerInjectionPercent  float64
	AverageLoanRatePercent float64
	BalanceSheetBalances   bool
}

// ProfilePage is the data for the profile page. The same page/template
// serves the plain display view, the name-edit form (Editing), and the
// delete-account error case (DeleteError) - see BuildProfilePage/
// BuildProfileEditPage/BuildProfilePageWithDeleteError.
type ProfilePage struct {
	Title        string
	User         *domain.User
	Editing      bool
	ErrorMessage string
	FirstName    string
	LastName     string
	DeleteError  string
}

// ErrorPage is the data for error pages
type ErrorPage struct {
	ErrorTitle       string
	ErrorStatusCode  int
	ErrorDescription string
	Message          string
	Path             string
}

// LoginPage is the data for the login page
type LoginPage struct {
	Title    string
	ErrorMsg string
	Email    string
}

// SignupPage is the data for the signup page
type SignupPage struct {
	Title     string
	ErrorMsg  string
	Email     string
	FirstName string
	LastName  string
}
