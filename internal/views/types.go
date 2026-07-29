package views

import (
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// BasePage contains common data for all pages
type BasePage struct {
	Title string
	Path  string
	User  *domain.User
}

// IndexPage is the data for the home page
type IndexPage struct {
	BasePage
	Plans []*domain.Plan
}

// SetupPage is the data for the plan setup page
type SetupPage struct {
	BasePage
	Plan *domain.Plan
}

// StartingPointPage is the data for the starting point page
type StartingPointPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

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
}

// BalanceSheetPage is the data for the balance sheet page
type BalanceSheetPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
}

// AnalyticsPage is the data for the analytics page
type AnalyticsPage struct {
	BasePage
	Plan         *domain.Plan
	ErrorMessage string
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
	Title    string
	ErrorMsg string
	Email    string
}
