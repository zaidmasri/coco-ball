// Package views handles types for templates
package views

import (
	"html/template"
	"log"
	"net/http"

	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// BuildBasePage creates a BasePage from request context and user
func BuildBasePage(r *http.Request, title string, user *domain.User) BasePage {
	return BasePage{
		Title: title,
		Path:  r.URL.Path,
		User:  user,
	}
}

// BuildIndexPage creates an IndexPage
func BuildIndexPage(r *http.Request, user *domain.User, plans []*domain.Plan, pendingInvites []InviteSummary) IndexPage {
	base := BuildBasePage(r, "Business Planning Tool", user)
	return IndexPage{
		BasePage:       base,
		Plans:          plans,
		PendingInvites: pendingInvites,
	}
}

// BuildSetupPage creates a SetupPage
func BuildSetupPage(r *http.Request, user *domain.User, plan *domain.Plan, invites []*domain.PlanInvite, isOwner bool) SetupPage {
	base := BuildBasePage(r, "Edit Setup | Business Planning Tool", user)
	return SetupPage{
		BasePage: base,
		Plan:     plan,
		Invites:  invites,
		IsOwner:  isOwner,
	}
}

// Starting Point's builders/renderers live in starting_point.go.

// BuildPayrollPage creates a PayrollPage
func BuildPayrollPage(r *http.Request, user *domain.User, plan *domain.Plan) PayrollPage {
	base := BuildBasePage(r, "Payroll | Business Planning Tool", user)
	return PayrollPage{
		BasePage: base,
		Plan:     plan,
	}
}

// BuildSalesForecastPage creates a SalesForecastPage
func BuildSalesForecastPage(r *http.Request, user *domain.User, plan *domain.Plan) SalesForecastPage {
	base := BuildBasePage(r, "Sales Forecast | Business Planning Tool", user)
	return SalesForecastPage{
		BasePage: base,
		Plan:     plan,
	}
}

// BuildOpExpensesPage creates an OpExpensesPage
func BuildOpExpensesPage(r *http.Request, user *domain.User, plan *domain.Plan) OpExpensesPage {
	base := BuildBasePage(r, "Operating Expenses | Business Planning Tool", user)
	return OpExpensesPage{
		BasePage: base,
		Plan:     plan,
	}
}

// BuildCashFlowPage creates a CashFlowPage
func BuildCashFlowPage(r *http.Request, user *domain.User, plan *domain.Plan) CashFlowPage {
	base := BuildBasePage(r, "Cash Flow | Business Planning Tool", user)
	return CashFlowPage{
		BasePage: base,
		Plan:     plan,
	}
}

// BuildIncomeStatementPage creates an IncomeStatementPage
func BuildIncomeStatementPage(r *http.Request, user *domain.User, plan *domain.Plan) IncomeStatementPage {
	base := BuildBasePage(r, "Income Statement | Business Planning Tool", user)
	return IncomeStatementPage{
		BasePage:  base,
		Plan:      plan,
		Years:     plan.AnnualSummaries(domain.ProjectionYears),
		Products:  plan.ProductFinancialsSeries(domain.ProjectionYears),
		OpExLines: plan.OpExAnnualBreakdown(domain.ProjectionYears),
	}
}

// BuildBalanceSheetPage creates a BalanceSheetPage
func BuildBalanceSheetPage(r *http.Request, user *domain.User, plan *domain.Plan) BalanceSheetPage {
	base := BuildBasePage(r, "Balance Sheet | Business Planning Tool", user)
	years := plan.BalanceSheetSnapshots(domain.ProjectionYears)

	balances := true
	for _, y := range years {
		diff := y.TotalAssets - y.TotalLiabilitiesAndEquity
		if diff < -25 || diff > 25 {
			balances = false
			break
		}
	}

	return BalanceSheetPage{
		BasePage: base,
		Plan:     plan,
		Years:    years,
		Balances: balances,
	}
}

// BuildAnalyticsPage creates an AnalyticsPage
func BuildAnalyticsPage(r *http.Request, user *domain.User, plan *domain.Plan) AnalyticsPage {
	base := BuildBasePage(r, "Analytics | Business Planning Tool", user)

	years := plan.BalanceSheetSnapshots(domain.ProjectionYears)
	balances := true
	for _, y := range years {
		diff := y.TotalAssets - y.TotalLiabilitiesAndEquity
		if diff < -25 || diff > 25 {
			balances = false
			break
		}
	}

	equity, debt := plan.FundingBreakdown()
	var ownerInjectionPercent float64
	if equity+debt != 0 {
		ownerInjectionPercent = float64(equity) / float64(equity+debt) * 100
	}

	return AnalyticsPage{
		BasePage:          base,
		Plan:              plan,
		Breakeven:         plan.Breakeven(),
		Ratios:            plan.FinancialRatiosSeries(domain.ProjectionYears),
		Years:             years,
		Annuals:           plan.AnnualSummaries(domain.ProjectionYears),
		AssetDepreciation: plan.AssetDepreciationBreakdown(),
		LoanAmortization:  plan.LoanAmortizationSummary(domain.ProjectionYears),

		OwnerInjectionPercent:  ownerInjectionPercent,
		AverageLoanRatePercent: plan.AverageLoanInterestRate() * 100,
		BalanceSheetBalances:   balances,
	}
}

// BuildProfilePage creates a ProfilePage
func BuildProfilePage(user *domain.User) ProfilePage {
	return ProfilePage{
		Title: "Profile | Business Planning Tool",
		User:  user,
	}
}

// BuildErrorPage creates an ErrorPage
func BuildErrorPage(r *http.Request, statusCode int, message string) ErrorPage {
	statusText := http.StatusText(statusCode)
	return ErrorPage{
		ErrorTitle:       statusText,
		ErrorStatusCode:  statusCode,
		ErrorDescription: statusText,
		Message:          message,
		Path:             r.URL.Path,
	}
}

// BuildLoginPage creates a LoginPage
func BuildLoginPage() LoginPage {
	return LoginPage{
		Title: "Login | Business Planning Tool",
	}
}

// BuildLoginPageWithError creates a LoginPage with an error message
func BuildLoginPageWithError(email, errMsg string) LoginPage {
	return LoginPage{
		Title:    "Login | Business Planning Tool",
		ErrorMsg: errMsg,
		Email:    email,
	}
}

// BuildSignupPage creates a SignupPage
func BuildSignupPage() SignupPage {
	return SignupPage{
		Title: "Sign Up | Business Planning Tool",
	}
}

// BuildSignupPageWithError creates a SignupPage with an error message,
// preserving whatever the user had already typed in.
func BuildSignupPageWithError(email, firstName, lastName, errMsg string) SignupPage {
	return SignupPage{
		Title:     "Sign Up | Business Planning Tool",
		ErrorMsg:  errMsg,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
}

// renderTemplate executes a template and writes to response writer
func renderTemplate(w http.ResponseWriter, tc map[string]*template.Template, cacheKey, templateName string, data any) {
	renderTemplateWithStatus(w, tc, cacheKey, templateName, data, http.StatusOK)
}

// renderTemplateWithStatus executes a template with a custom status code
func renderTemplateWithStatus(w http.ResponseWriter, tc map[string]*template.Template, cacheKey, templateName string, data any, statusCode int) {
	ts, ok := tc[cacheKey]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if err := ts.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("❌ Template Execution Error (%s): %v", cacheKey, err)
	}
}

// RenderIndexPage renders the index page
func RenderIndexPage(w http.ResponseWriter, tc map[string]*template.Template, page IndexPage) {
	renderTemplate(w, tc, "index.html", "index.html", page)
}

// RenderSetupPage renders the setup page
func RenderSetupPage(w http.ResponseWriter, tc map[string]*template.Template, page SetupPage) {
	renderTemplate(w, tc, "setup.html", "base", page)
}

// RenderPayrollPage renders the payroll page
func RenderPayrollPage(w http.ResponseWriter, tc map[string]*template.Template, page PayrollPage) {
	renderTemplate(w, tc, "payroll.html", "base", page)
}

// RenderSalesForecastPage renders the sales forecast page
func RenderSalesForecastPage(w http.ResponseWriter, tc map[string]*template.Template, page SalesForecastPage) {
	renderTemplate(w, tc, "sales-forecast.html", "base", page)
}

// RenderOpExpensesPage renders the operating expenses page
func RenderOpExpensesPage(w http.ResponseWriter, tc map[string]*template.Template, page OpExpensesPage) {
	renderTemplate(w, tc, "op-expenses.html", "base", page)
}

// RenderCashFlowPage renders the cash flow page
func RenderCashFlowPage(w http.ResponseWriter, tc map[string]*template.Template, page CashFlowPage) {
	renderTemplate(w, tc, "cash-flow.html", "base", page)
}

// RenderIncomeStatementPage renders the income statement page
func RenderIncomeStatementPage(w http.ResponseWriter, tc map[string]*template.Template, page IncomeStatementPage) {
	renderTemplate(w, tc, "income-statement.html", "base", page)
}

// RenderBalanceSheetPage renders the balance sheet page
func RenderBalanceSheetPage(w http.ResponseWriter, tc map[string]*template.Template, page BalanceSheetPage) {
	renderTemplate(w, tc, "balance-sheet.html", "base", page)
}

// RenderAnalyticsPage renders the analytics page
func RenderAnalyticsPage(w http.ResponseWriter, tc map[string]*template.Template, page AnalyticsPage) {
	renderTemplate(w, tc, "analytics.html", "base", page)
}

// RenderProfilePage renders the profile page
func RenderProfilePage(w http.ResponseWriter, tc map[string]*template.Template, page ProfilePage) {
	renderTemplate(w, tc, "profile.html", "profile.html", page)
}

// RenderErrorPage renders the error page
func RenderErrorPage(w http.ResponseWriter, tc map[string]*template.Template, page ErrorPage) {
	renderTemplate(w, tc, "error.html", "error.html", page)
}

// RenderErrorPageWithStatus renders the error page with a specific status code
func RenderErrorPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page ErrorPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "error.html", "error.html", page, statusCode)
}

// RenderLoginPage renders the login page
func RenderLoginPage(w http.ResponseWriter, tc map[string]*template.Template, page LoginPage) {
	renderTemplate(w, tc, "login.html", "login.html", page)
}

// RenderSignupPage renders the signup page
func RenderSignupPage(w http.ResponseWriter, tc map[string]*template.Template, page SignupPage) {
	renderTemplate(w, tc, "signup.html", "signup.html", page)
}
