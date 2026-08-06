// Package views - Operating Expenses wizard. Unlike Payroll/Sales
// Forecast/Cash Flow, Operating Expenses is a single-section hub: one
// repeatable list of expense line items, with no multi-section summary
// page to route through first - the hub's "home" is the list page itself,
// the same way Fixed Assets/Salary Roles/Products are one repeatable
// section within a larger hub. Built on the same shared wizard
// infrastructure (SectionListPage, AddAnotherPage, QuestionStepPage)
// starting_point.go introduced.
package views

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// IsOperatingExpensesComplete reports whether the Operating Expenses
// section is complete, given the section-status map the store returns.
// Used to decide whether the sidebar's Operating Expenses nav icon
// renders filled.
func IsOperatingExpensesComplete(sectionStatus map[string]bool) bool {
	return sectionStatus[domain.SectionOperatingExpenses]
}

// --- Answered-so-far summaries ---

var opExpenseFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.Cost) string
}{
	{"name", "Name", func(c domain.Cost) string { return c.Name }},
	{"amount", "Monthly Amount", func(c domain.Cost) string { return formatMoney(c.BaseAmountPerMonth) }},
	{"growth", "Annual Growth", func(c domain.Cost) string { return formatPercent(c.Growth.AnnualRatePercent()) }},
}

func opExpenseAnsweredFields(cost domain.Cost, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range opExpenseFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(cost)})
	}
	return out
}

// --- Suggestion pills ---

var opExpenseSuggestions = []string{
	"Advertising & Marketing", "Car and Truck Expenses", "Commissions and Fees",
	"Contract Labor (1099)", "Insurance (non-health)", "Legal and Professional Services",
	"Licenses and Permits", "Office Expense", "Rent/Lease (Vehicles & Equipment)",
	"Rent/Lease (Real Estate)", "Repairs and Maintenance", "Supplies",
	"Travel, Meals, and Entertainment", "Utilities", "Miscellaneous",
}

// --- URL helpers ---

// OperatingExpensesListURL is the Operating Expenses section's list/entry
// page - this hub's "home", since it has no multi-section summary page.
func OperatingExpensesListURL(planID uuid.UUID) string {
	return fmt.Sprintf("/plan/%s/operating-expenses", planID)
}

// OperatingExpensesStepURL is one wizard question for a specific
// Operating Expense item.
func OperatingExpensesStepURL(planID, itemID uuid.UUID, step string) string {
	return fmt.Sprintf("/plan/%s/operating-expenses/%s/%s", planID, itemID, step)
}

// --- Page types ---

// OpExpenseItem pairs an Operating Expense's own wizard-item ID with its
// underlying domain value, so the list page can render Edit/Delete links
// without Cost itself needing to know about wizard/storage identity.
type OpExpenseItem struct {
	ID   uuid.UUID
	Cost domain.Cost
}

// OpExpenseStepPage renders a single wizard question for one Operating
// Expense item.
type OpExpenseStepPage struct {
	QuestionStepPage
	Cost domain.Cost // answers so far, for pre-fill on Back/error
}

// --- Builders ---

// opExpenseListItem formats an Operating Expense for the shared
// section-list page.
func opExpenseListItem(planID uuid.UUID, item OpExpenseItem) SectionListItem {
	detail := formatMoney(item.Cost.BaseAmountPerMonth) + "/mo"
	if item.Cost.Growth.Type == domain.AnnualStepPercent && item.Cost.Growth.AnnualRate != 0 {
		detail += fmt.Sprintf(" · %s/yr growth", formatPercent(item.Cost.Growth.AnnualRatePercent()))
	}
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Cost.Name,
		Detail:       detail,
		EditURL:      OperatingExpensesStepURL(planID, item.ID, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/operating-expenses/%s", planID, item.ID),
	}
}

// BuildOperatingExpenseListPage creates the Operating Expenses section's
// list page.
func BuildOperatingExpenseListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []OpExpenseItem, complete bool, draftItemID *uuid.UUID, draftStep string, operatingExpensesComplete bool) SectionListPage {
	base := BuildBasePage(r, "Operating Expenses | Business Planning Tool", user)
	base.OperatingExpensesComplete = operatingExpensesComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = opExpenseListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = OperatingExpensesStepURL(plan.ID(), *draftItemID, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Step 5 of 9",
		SectionIcon:        "bi-receipt",
		SectionTitle:       "Operating Expenses",
		SectionDescription: "Let's map out your ongoing fixed overhead. Enter your expected monthly expenses and an estimated annual inflation or growth rate for each.",
		EmptyIcon:          "bi-tags",
		EmptyText:          "No operating expenses added yet.",
		ItemNounSingular:   "operating expense",
		AddButtonLabel:     "Add Expense",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/operating-expenses/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/operating-expenses/finish", plan.ID()),
		OverviewURL:        SalesForecastSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildOperatingExpenseStepPage creates an OpExpenseStepPage.
func BuildOperatingExpenseStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, cost domain.Cost, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, operatingExpensesComplete bool) OpExpenseStepPage {
	base := BuildBasePage(r, "Operating Expenses | Business Planning Tool", user)
	base.OperatingExpensesComplete = operatingExpensesComplete
	return OpExpenseStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    "Operating Expenses",
			SectionIcon:     "bi-receipt",
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      OperatingExpensesStepURL(plan.ID(), itemID, step),
			BackURL:         backURL,
			CancelURL:       OperatingExpensesListURL(plan.ID()),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: opExpenseAnsweredFields(cost, step),
			Suggestions:     suggestionsForStep(step, "name", opExpenseSuggestions),
		},
		Cost: cost,
	}
}

// BuildOperatingExpenseAddAnotherPage creates the Operating Expenses "add
// another?" interstitial.
func BuildOperatingExpenseAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, cost domain.Cost, operatingExpensesComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Operating Expenses | Business Planning Tool", user)
	base.OperatingExpensesComplete = operatingExpensesComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         cost.Name,
		Answers:          opExpenseAnsweredFields(cost, ""),
		ItemNounSingular: "operating expense",
		DoneAction:       fmt.Sprintf("/plan/%s/operating-expenses/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/operating-expenses/new", plan.ID()),
	}
}

// --- Renderers ---

// RenderOperatingExpenseStepPage renders a single Operating Expense
// wizard step.
func RenderOperatingExpenseStepPage(w http.ResponseWriter, tc map[string]*template.Template, page OpExpenseStepPage) {
	renderTemplate(w, tc, "operating-expenses-step.html", "base", page)
}

// RenderOperatingExpenseStepPageWithStatus renders an Operating Expense
// wizard step with a custom status code (used on validation errors).
func RenderOperatingExpenseStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page OpExpenseStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "operating-expenses-step.html", "base", page, statusCode)
}
