package views

import (
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// BasePage contains common data for all pages
type BasePage struct {
	Title string
	Path  string
	User  *domain.User

	// StartingPointComplete reports whether every Starting Point
	// sub-section (Fixed Assets, Startup Costs, Funding Sources, Cash on
	// Hand) is complete. The sidebar (rendered on every page) uses it to
	// decide whether the Starting Point nav icon shows filled or outline.
	StartingPointComplete bool
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

// Starting Point's page types, builders, and renderers live in
// starting_point.go, not here - they're numerous enough (summary, 3
// repeatable sections x list/step/add-another, Cash on Hand, section
// intro) to warrant their own file rather than growing this one.

// PayrollPage is the data for the payroll page
type PayrollPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

// SalesForecastPage is the data for the sales forecast page
type SalesForecastPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

// OpExpensesPage is the data for the operating expenses page
type OpExpensesPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

// CashFlowPage is the data for the cash flow page
type CashFlowPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

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

// ProfilePage is the data for the profile page
type ProfilePage struct {
	Title string
	User  *domain.User
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
