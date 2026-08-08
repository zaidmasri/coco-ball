// Package views - Payroll hub: a summary page (/plan/{id}/payroll) listing
// Salary Roles, Benefits, and Payroll Tax Rates completion status, plus one
// wizard per section, all built on the same shared wizard infrastructure
// (SectionListPage, AddAnotherPage, SectionIntroPage, HubSummaryPage,
// QuestionStepPage) Starting Point introduced in wizard_shared.go and
// starting_point.go.
package views

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// --- Section metadata ---

var payrollSectionMetas = []sectionMeta{
	{
		Key:       domain.SectionSalaryRoles,
		Title:     "Salary Roles",
		Icon:      "bi-people",
		ShortDesc: "The roles/titles on your team and what they're paid.",
		LongDesc: "Salary Roles are the people (or contractors) your business pays " +
			"regularly - a title, headcount, and monthly pay for each. We'll walk " +
			"through them one at a time, including how much that role's pay grows " +
			"in Year 2 and Year 3.",
	},
	{
		Key:       domain.SectionBenefits,
		Title:     "Benefits",
		Icon:      "bi-heart-pulse",
		ShortDesc: "Health insurance, retirement, and other benefits you offer.",
		LongDesc: "Benefits are the recurring costs of employing people beyond " +
			"salary - health insurance, retirement contributions, workers' " +
			"compensation, and similar. We'll capture each one's monthly cost and " +
			"how it grows over time.",
	},
	{
		Key:       domain.SectionPayrollTaxRates,
		Title:     "Payroll Tax Rates",
		Icon:      "bi-bank",
		ShortDesc: "Employer-side Social Security, Medicare, and unemployment rates.",
		LongDesc: "Payroll Tax Rates are the employer-side taxes applied against " +
			"your W-2 payroll - Social Security, Medicare, and federal/state " +
			"unemployment. These are standard rates most businesses can use as-is.",
	},
}

func payrollSectionMetaByKey(key string) sectionMeta {
	return sectionMetaByKey(payrollSectionMetas, key)
}

// IsPayrollComplete reports whether every Payroll sub-section is complete,
// given the section-status map the store returns. Used to decide whether
// the sidebar's Payroll nav icon renders filled.
func IsPayrollComplete(sectionStatus map[string]bool) bool {
	return hubComplete(sectionStatus, payrollSectionMetas)
}

// --- Answered-so-far summaries ---

var salaryRoleFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.SalaryRole) string
}{
	{"role", "Role", func(s domain.SalaryRole) string { return s.Role }},
	{"type", "Type", func(s domain.SalaryRole) string {
		if s.IsContractor {
			return "1099 Contractor"
		}
		return "Employee"
	}},
	{"headcount", "Headcount", func(s domain.SalaryRole) string { return fmt.Sprintf("%d", s.Headcount) }},
	{"monthly-pay", "Monthly Pay", func(s domain.SalaryRole) string { return formatMoney(s.MonthlyPay) }},
	{"growth-yr2", "Year 2 Growth", func(s domain.SalaryRole) string { return formatPercent(s.GrowthRatePercent(0)) }},
	{"growth-yr3", "Year 3 Growth", func(s domain.SalaryRole) string { return formatPercent(s.GrowthRatePercent(1)) }},
}

func salaryRoleAnsweredFields(role domain.SalaryRole, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range salaryRoleFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(role)})
	}
	return out
}

var benefitFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.Benefit) string
}{
	{"type", "Benefit Type", func(b domain.Benefit) string { return b.Type }},
	{"monthly-amount", "Monthly Cost", func(b domain.Benefit) string { return formatMoney(b.MonthlyAmount) }},
	{"growth-yr2", "Year 2 Growth", func(b domain.Benefit) string { return formatPercent(b.GrowthRatePercent(0)) }},
	{"growth-yr3", "Year 3 Growth", func(b domain.Benefit) string { return formatPercent(b.GrowthRatePercent(1)) }},
}

func benefitAnsweredFields(benefit domain.Benefit, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range benefitFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(benefit)})
	}
	return out
}

var payrollTaxRatesFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.PayrollTaxRates) string
}{
	{"social-security", "Social Security", func(r domain.PayrollTaxRates) string { return formatPercent(r.SocialSecurityRatePercent()) }},
	{"medicare", "Medicare", func(r domain.PayrollTaxRates) string { return formatPercent(r.MedicareRatePercent()) }},
	{"futa", "FUTA", func(r domain.PayrollTaxRates) string { return formatPercent(r.FUTARatePercent()) }},
	{"suta", "SUTA", func(r domain.PayrollTaxRates) string { return formatPercent(r.SUTARatePercent()) }},
}

func payrollTaxRatesAnsweredFields(rates domain.PayrollTaxRates, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range payrollTaxRatesFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(rates)})
	}
	return out
}

// --- Suggestion pills ---

var (
	salaryRoleSuggestions = []string{"Owner/Founder", "Full-Time Employees", "Part-Time Employees", "Independent Contractors"}
	benefitSuggestions    = []string{"Employee Health Insurance", "Employee Pension/401k", "Worker's Compensation", "Other Benefit Programs"}
)

// --- URL helpers ---

// PayrollSummaryURL is the Payroll overview page for a plan.
func PayrollSummaryURL(planID uuid.UUID) string {
	return fmt.Sprintf("/plan/%s/payroll", planID)
}

// PayrollSectionListURL is a repeatable Payroll section's list/entry page.
func PayrollSectionListURL(planID uuid.UUID, section string) string {
	return fmt.Sprintf("/plan/%s/payroll/%s", planID, section)
}

// PayrollSectionStepURL is one wizard question for a specific item in a
// repeatable Payroll section (Salary Roles, Benefits).
func PayrollSectionStepURL(planID, itemID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/payroll/%s/%s/%s", planID, section, itemID, step)
}

// PayrollSectionSingletonStepURL is one wizard question for a singleton
// Payroll section (Payroll Tax Rates), which has no item ID.
func PayrollSectionSingletonStepURL(planID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/payroll/%s/%s", planID, section, step)
}

// --- Page types ---

// SalaryRoleItem pairs a Salary Role's own wizard-item ID with its
// underlying domain value, so the list page can render Edit/Delete links
// without SalaryRole itself needing to know about wizard/storage identity.
type SalaryRoleItem struct {
	ID   uuid.UUID
	Role domain.SalaryRole
}

// BenefitItem is the Benefits equivalent of SalaryRoleItem.
type BenefitItem struct {
	ID      uuid.UUID
	Benefit domain.Benefit
}

// SalaryRoleStepPage renders a single wizard question for one Salary Role
// item.
type SalaryRoleStepPage struct {
	QuestionStepPage
	Role domain.SalaryRole // answers so far, for pre-fill on Back/error
}

// BenefitStepPage is the Benefits equivalent of SalaryRoleStepPage.
type BenefitStepPage struct {
	QuestionStepPage
	Benefit domain.Benefit
}

// PayrollTaxRatesStepPage renders a single wizard question for the plan's
// Payroll Tax Rates singleton.
type PayrollTaxRatesStepPage struct {
	QuestionStepPage
	Rates domain.PayrollTaxRates
}

// --- Builders ---

// BuildPayrollSummaryPage creates the Payroll overview page.
func BuildPayrollSummaryPage(r *http.Request, user *domain.User, plan *domain.Plan, sectionStatus map[string]bool, salaryRoleCount, benefitCount int) HubSummaryPage {
	base := BuildBasePage(r, "Payroll | Business Planning Tool", user)
	base.PayrollComplete = IsPayrollComplete(sectionStatus)

	counts := map[string]int{
		domain.SectionSalaryRoles: salaryRoleCount,
		domain.SectionBenefits:    benefitCount,
	}

	sections := make([]HubSectionStatus, len(payrollSectionMetas))
	for i, m := range payrollSectionMetas {
		sections[i] = HubSectionStatus{
			Key:           m.Key,
			Title:         m.Title,
			Description:   m.ShortDesc,
			Complete:      sectionStatus[m.Key],
			ItemCount:     counts[m.Key],
			EditURL:       PayrollSectionListURL(plan.ID(), m.Key),
			GetStartedURL: fmt.Sprintf("/plan/%s/payroll/intro/%s", plan.ID(), m.Key),
		}
	}

	return HubSummaryPage{
		BasePage:       base,
		Plan:           plan,
		StepBadge:      "Step 3 of 9",
		HubTitle:       "Payroll & Benefits",
		HubDescription: "Let's map out your team. Define your personnel costs, baseline payroll taxes, and any benefits you plan to offer.",
		Sections:       sections,
		BackURL:        StartingPointSummaryURL(plan.ID()),
		ContinueURL:    fmt.Sprintf("/plan/%s/sales-forecast", plan.ID()),
	}
}

// BuildPayrollSectionIntroPage creates the landing/description page shown
// when a user clicks "Get Started" on a Payroll section. section must be
// one of the domain.SectionSalaryRoles/Benefits/PayrollTaxRates constants.
func BuildPayrollSectionIntroPage(r *http.Request, user *domain.User, plan *domain.Plan, section string, payrollComplete bool) SectionIntroPage {
	meta := payrollSectionMetaByKey(section)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete

	return SectionIntroPage{
		BasePage:      base,
		Plan:          plan,
		HubBadge:      "Payroll",
		HubSummaryURL: PayrollSummaryURL(plan.ID()),
		SectionTitle:  meta.Title,
		SectionIcon:   meta.Icon,
		Description:   meta.LongDesc,
		ContinueURL:   PayrollSectionListURL(plan.ID(), section),
	}
}

// salaryRoleListItem formats a Salary Role for the shared section-list page.
func salaryRoleListItem(planID uuid.UUID, item SalaryRoleItem) SectionListItem {
	kind := "Employee"
	if item.Role.IsContractor {
		kind = "1099 Contractor"
	}
	detail := fmt.Sprintf("%s/mo · %d headcount · %s", formatMoney(item.Role.MonthlyPay), item.Role.Headcount, kind)
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Role.Role,
		Detail:       detail,
		EditURL:      PayrollSectionStepURL(planID, item.ID, domain.SectionSalaryRoles, "role"),
		DeleteAction: fmt.Sprintf("/plan/%s/payroll/salary-roles/%s", planID, item.ID),
	}
}

// BuildSalaryRoleListPage creates the Salary Roles section's list page.
func BuildSalaryRoleListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []SalaryRoleItem, complete bool, draftItemID *uuid.UUID, draftStep string, payrollComplete bool) SectionListPage {
	base := BuildBasePage(r, "Salary Roles | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = salaryRoleListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = PayrollSectionStepURL(plan.ID(), *draftItemID, domain.SectionSalaryRoles, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Payroll",
		SectionIcon:        "bi-people",
		SectionTitle:       "Salary Roles",
		SectionDescription: "The roles/titles on your team and what they're paid.",
		EmptyIcon:          "bi-person-add",
		EmptyText:          "No salary roles yet.",
		ItemNounSingular:   "salary role",
		AddButtonLabel:     "Add Role",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/payroll/salary-roles/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/payroll/salary-roles/finish", plan.ID()),
		OverviewURL:        PayrollSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildSalaryRoleStepPage creates a SalaryRoleStepPage.
func BuildSalaryRoleStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, role domain.SalaryRole, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, payrollComplete bool) SalaryRoleStepPage {
	meta := payrollSectionMetaByKey(domain.SectionSalaryRoles)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete
	return SalaryRoleStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      PayrollSectionStepURL(plan.ID(), itemID, domain.SectionSalaryRoles, step),
			BackURL:         backURL,
			CancelURL:       PayrollSectionListURL(plan.ID(), domain.SectionSalaryRoles),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: salaryRoleAnsweredFields(role, step),
			Suggestions:     suggestionsForStep(step, "role", salaryRoleSuggestions),
		},
		Role: role,
	}
}

// BuildSalaryRoleAddAnotherPage creates the Salary Roles "add another?"
// interstitial.
func BuildSalaryRoleAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, role domain.SalaryRole, payrollComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Salary Roles | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         role.Role,
		Answers:          salaryRoleAnsweredFields(role, ""),
		ItemNounSingular: "salary role",
		DoneAction:       fmt.Sprintf("/plan/%s/payroll/salary-roles/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/payroll/salary-roles/new", plan.ID()),
	}
}

// benefitListItem formats a Benefit for the shared section-list page.
func benefitListItem(planID uuid.UUID, item BenefitItem) SectionListItem {
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Benefit.Type,
		Detail:       formatMoney(item.Benefit.MonthlyAmount) + "/mo",
		EditURL:      PayrollSectionStepURL(planID, item.ID, domain.SectionBenefits, "type"),
		DeleteAction: fmt.Sprintf("/plan/%s/payroll/benefits/%s", planID, item.ID),
	}
}

// BuildBenefitListPage creates the Benefits section's list page.
func BuildBenefitListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []BenefitItem, complete bool, draftItemID *uuid.UUID, draftStep string, payrollComplete bool) SectionListPage {
	base := BuildBasePage(r, "Benefits | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = benefitListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = PayrollSectionStepURL(plan.ID(), *draftItemID, domain.SectionBenefits, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Payroll",
		SectionIcon:        "bi-heart-pulse",
		SectionTitle:       "Benefits",
		SectionDescription: "Health insurance, retirement, and other benefits you offer.",
		EmptyIcon:          "bi-shield-plus",
		EmptyText:          "No benefits yet.",
		ItemNounSingular:   "benefit",
		AddButtonLabel:     "Add Benefit",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/payroll/benefits/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/payroll/benefits/finish", plan.ID()),
		OverviewURL:        PayrollSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildBenefitStepPage creates a BenefitStepPage.
func BuildBenefitStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, benefit domain.Benefit, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, payrollComplete bool) BenefitStepPage {
	meta := payrollSectionMetaByKey(domain.SectionBenefits)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete
	return BenefitStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      PayrollSectionStepURL(plan.ID(), itemID, domain.SectionBenefits, step),
			BackURL:         backURL,
			CancelURL:       PayrollSectionListURL(plan.ID(), domain.SectionBenefits),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: benefitAnsweredFields(benefit, step),
			Suggestions:     suggestionsForStep(step, "type", benefitSuggestions),
		},
		Benefit: benefit,
	}
}

// BuildBenefitAddAnotherPage creates the Benefits "add another?" interstitial.
func BuildBenefitAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, benefit domain.Benefit, payrollComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Benefits | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         benefit.Type,
		Answers:          benefitAnsweredFields(benefit, ""),
		ItemNounSingular: "benefit",
		DoneAction:       fmt.Sprintf("/plan/%s/payroll/benefits/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/payroll/benefits/new", plan.ID()),
	}
}

// BuildPayrollTaxRatesStepPage creates a PayrollTaxRatesStepPage.
func BuildPayrollTaxRatesStepPage(r *http.Request, user *domain.User, plan *domain.Plan, rates domain.PayrollTaxRates, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, payrollComplete bool) PayrollTaxRatesStepPage {
	meta := payrollSectionMetaByKey(domain.SectionPayrollTaxRates)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.PayrollComplete = payrollComplete
	return PayrollTaxRatesStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      PayrollSectionSingletonStepURL(plan.ID(), domain.SectionPayrollTaxRates, step),
			BackURL:         backURL,
			CancelURL:       PayrollSummaryURL(plan.ID()), // Payroll Tax Rates has no list page to cancel back to
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: payrollTaxRatesAnsweredFields(rates, step),
		},
		Rates: rates,
	}
}

// --- Renderers ---

// RenderSalaryRoleStepPage renders a single Salary Role wizard step.
func RenderSalaryRoleStepPage(w http.ResponseWriter, tc map[string]*template.Template, page SalaryRoleStepPage) {
	renderTemplate(w, tc, "payroll-salary-roles-step.html", "base", page)
}

// RenderSalaryRoleStepPageWithStatus renders a Salary Role wizard step with
// a custom status code (used on validation errors).
func RenderSalaryRoleStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page SalaryRoleStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "payroll-salary-roles-step.html", "base", page, statusCode)
}

// RenderBenefitStepPage renders a single Benefit wizard step.
func RenderBenefitStepPage(w http.ResponseWriter, tc map[string]*template.Template, page BenefitStepPage) {
	renderTemplate(w, tc, "payroll-benefits-step.html", "base", page)
}

// RenderBenefitStepPageWithStatus renders a Benefit wizard step with a
// custom status code (used on validation errors).
func RenderBenefitStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page BenefitStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "payroll-benefits-step.html", "base", page, statusCode)
}

// RenderPayrollTaxRatesStepPage renders a single Payroll Tax Rates wizard
// step.
func RenderPayrollTaxRatesStepPage(w http.ResponseWriter, tc map[string]*template.Template, page PayrollTaxRatesStepPage) {
	renderTemplate(w, tc, "payroll-tax-rates-step.html", "base", page)
}

// RenderPayrollTaxRatesStepPageWithStatus renders a Payroll Tax Rates
// wizard step with a custom status code (used on validation errors).
func RenderPayrollTaxRatesStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page PayrollTaxRatesStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "payroll-tax-rates-step.html", "base", page, statusCode)
}
