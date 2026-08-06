// Package views - Cash Flow hub: a summary page (/plan/{id}/cash-flow)
// listing Inventory Purchases and Distributions completion status, plus
// one wizard per section, built on the same shared wizard infrastructure
// Starting Point introduced.
package views

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// --- Section metadata ---

var cashFlowSectionMetas = []sectionMeta{
	{
		Key:       domain.SectionInventoryPurchases,
		Title:     "Inventory Purchases",
		Icon:      "bi-boxes",
		ShortDesc: "Additional inventory stockpiling independent of monthly COGS.",
		LongDesc: "Inventory Purchases are discretionary cash outflows for " +
			"stockpiling inventory beyond what's already captured in your monthly " +
			"COGS - useful if you buy ahead of a busy season, for example.",
	},
	{
		Key:       domain.SectionDistributions,
		Title:     "Distributions",
		Icon:      "bi-cash-stack",
		ShortDesc: "Owner draws, dividends, and loan/credit repayments.",
		LongDesc: "Distributions are discretionary cash outflows to owners or " +
			"lenders - draws, dividends, or line-of-credit repayments - that " +
			"aren't part of your standard operating expenses.",
	},
}

func cashFlowSectionMetaByKey(key string) sectionMeta {
	return sectionMetaByKey(cashFlowSectionMetas, key)
}

// IsCashFlowComplete reports whether every Cash Flow sub-section is
// complete, given the section-status map the store returns. Used to
// decide whether the sidebar's Cash Flow nav icon renders filled.
func IsCashFlowComplete(sectionStatus map[string]bool) bool {
	return hubComplete(sectionStatus, cashFlowSectionMetas)
}

// --- Answered-so-far summaries ---

var inventoryPurchaseFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.InventoryPurchase) string
}{
	{"category", "Category", func(i domain.InventoryPurchase) string { return i.Category }},
	{"monthly-amount", "Monthly Amount", func(i domain.InventoryPurchase) string { return formatMoney(i.MonthlyAmount) }},
	{"growth-yr2", "Year 2 Growth", func(i domain.InventoryPurchase) string { return formatPercent(i.GrowthRatePercent(0)) }},
	{"growth-yr3", "Year 3 Growth", func(i domain.InventoryPurchase) string { return formatPercent(i.GrowthRatePercent(1)) }},
}

func inventoryPurchaseAnsweredFields(inv domain.InventoryPurchase, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range inventoryPurchaseFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(inv)})
	}
	return out
}

var distributionFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.Distribution) string
}{
	{"name", "Name", func(d domain.Distribution) string { return d.Name }},
	{"monthly-amount", "Monthly Amount", func(d domain.Distribution) string { return formatMoney(d.MonthlyAmount) }},
	{"growth-yr2", "Year 2 Growth", func(d domain.Distribution) string { return formatPercent(d.GrowthRatePercent(0)) }},
	{"growth-yr3", "Year 3 Growth", func(d domain.Distribution) string { return formatPercent(d.GrowthRatePercent(1)) }},
}

func distributionAnsweredFields(dist domain.Distribution, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range distributionFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(dist)})
	}
	return out
}

// --- Suggestion pills ---

var (
	inventoryPurchaseSuggestions = []string{"Raw Materials", "Finished Goods", "Packaging & Shipping Supplies"}
	distributionSuggestions      = []string{"Owner's Distribution", "Dividends Paid", "Line of Credit Repayment", "Partner Draw"}
)

// --- URL helpers ---

// CashFlowSummaryURL is the Cash Flow overview page for a plan.
func CashFlowSummaryURL(planID uuid.UUID) string {
	return fmt.Sprintf("/plan/%s/cash-flow", planID)
}

// CashFlowSectionListURL is a repeatable Cash Flow section's list/entry page.
func CashFlowSectionListURL(planID uuid.UUID, section string) string {
	return fmt.Sprintf("/plan/%s/cash-flow/%s", planID, section)
}

// CashFlowSectionStepURL is one wizard question for a specific item in a
// repeatable Cash Flow section (Inventory Purchases, Distributions).
func CashFlowSectionStepURL(planID, itemID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/cash-flow/%s/%s/%s", planID, section, itemID, step)
}

// --- Page types ---

// InventoryPurchaseItem pairs an Inventory Purchase's own wizard-item ID
// with its underlying domain value, so the list page can render
// Edit/Delete links without InventoryPurchase itself needing to know about
// wizard/storage identity.
type InventoryPurchaseItem struct {
	ID       uuid.UUID
	Purchase domain.InventoryPurchase
}

// DistributionItem is the Distributions equivalent of InventoryPurchaseItem.
type DistributionItem struct {
	ID           uuid.UUID
	Distribution domain.Distribution
}

// InventoryPurchaseStepPage renders a single wizard question for one
// Inventory Purchase item.
type InventoryPurchaseStepPage struct {
	QuestionStepPage
	Purchase domain.InventoryPurchase // answers so far, for pre-fill on Back/error
}

// DistributionStepPage is the Distributions equivalent of
// InventoryPurchaseStepPage.
type DistributionStepPage struct {
	QuestionStepPage
	Distribution domain.Distribution
}

// --- Builders ---

// BuildCashFlowSummaryPage creates the Cash Flow overview page.
func BuildCashFlowSummaryPage(r *http.Request, user *domain.User, plan *domain.Plan, sectionStatus map[string]bool, inventoryCount, distributionCount int) HubSummaryPage {
	base := BuildBasePage(r, "Cash Flow | Business Planning Tool", user)
	base.CashFlowComplete = IsCashFlowComplete(sectionStatus)

	counts := map[string]int{
		domain.SectionInventoryPurchases: inventoryCount,
		domain.SectionDistributions:      distributionCount,
	}

	sections := make([]HubSectionStatus, len(cashFlowSectionMetas))
	for i, m := range cashFlowSectionMetas {
		sections[i] = HubSectionStatus{
			Key:           m.Key,
			Title:         m.Title,
			Description:   m.ShortDesc,
			Complete:      sectionStatus[m.Key],
			ItemCount:     counts[m.Key],
			EditURL:       CashFlowSectionListURL(plan.ID(), m.Key),
			GetStartedURL: fmt.Sprintf("/plan/%s/cash-flow/intro/%s", plan.ID(), m.Key),
		}
	}

	return HubSummaryPage{
		BasePage:       base,
		Plan:           plan,
		StepBadge:      "Step 6 of 9",
		HubTitle:       "Discretionary Cash Flow",
		HubDescription: "Let's account for specific cash outflows that aren't captured in your standard operating expenses, such as inventory stockpiling or paying owners.",
		Sections:       sections,
		BackURL:        fmt.Sprintf("/plan/%s/operating-expenses", plan.ID()),
		ContinueURL:    fmt.Sprintf("/plan/%s/income-statement", plan.ID()),
	}
}

// BuildCashFlowSectionIntroPage creates the landing/description page shown
// when a user clicks "Get Started" on a Cash Flow section. section must be
// one of the domain.SectionInventoryPurchases/Distributions constants.
func BuildCashFlowSectionIntroPage(r *http.Request, user *domain.User, plan *domain.Plan, section string, cashFlowComplete bool) SectionIntroPage {
	meta := cashFlowSectionMetaByKey(section)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete

	return SectionIntroPage{
		BasePage:      base,
		Plan:          plan,
		HubBadge:      "Cash Flow",
		HubSummaryURL: CashFlowSummaryURL(plan.ID()),
		SectionTitle:  meta.Title,
		SectionIcon:   meta.Icon,
		Description:   meta.LongDesc,
		ContinueURL:   CashFlowSectionListURL(plan.ID(), section),
	}
}

// inventoryPurchaseListItem formats an Inventory Purchase for the shared
// section-list page.
func inventoryPurchaseListItem(planID uuid.UUID, item InventoryPurchaseItem) SectionListItem {
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Purchase.Category,
		Detail:       formatMoney(item.Purchase.MonthlyAmount) + "/mo",
		EditURL:      CashFlowSectionStepURL(planID, item.ID, domain.SectionInventoryPurchases, "category"),
		DeleteAction: fmt.Sprintf("/plan/%s/cash-flow/inventory-purchases/%s", planID, item.ID),
	}
}

// BuildInventoryPurchaseListPage creates the Inventory Purchases section's
// list page.
func BuildInventoryPurchaseListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []InventoryPurchaseItem, complete bool, draftItemID *uuid.UUID, draftStep string, cashFlowComplete bool) SectionListPage {
	base := BuildBasePage(r, "Inventory Purchases | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = inventoryPurchaseListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = CashFlowSectionStepURL(plan.ID(), *draftItemID, domain.SectionInventoryPurchases, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Cash Flow",
		SectionIcon:        "bi-boxes",
		SectionTitle:       "Inventory Purchases",
		SectionDescription: "Additional inventory stockpiling independent of monthly COGS.",
		EmptyIcon:          "bi-box-seam",
		EmptyText:          "No additional inventory planned.",
		ItemNounSingular:   "inventory purchase",
		AddButtonLabel:     "Add Inventory",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/cash-flow/inventory-purchases/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/cash-flow/inventory-purchases/finish", plan.ID()),
		OverviewURL:        CashFlowSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildInventoryPurchaseStepPage creates an InventoryPurchaseStepPage.
func BuildInventoryPurchaseStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, purchase domain.InventoryPurchase, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, cashFlowComplete bool) InventoryPurchaseStepPage {
	meta := cashFlowSectionMetaByKey(domain.SectionInventoryPurchases)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete
	return InventoryPurchaseStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      CashFlowSectionStepURL(plan.ID(), itemID, domain.SectionInventoryPurchases, step),
			BackURL:         backURL,
			CancelURL:       CashFlowSectionListURL(plan.ID(), domain.SectionInventoryPurchases),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: inventoryPurchaseAnsweredFields(purchase, step),
			Suggestions:     suggestionsForStep(step, "category", inventoryPurchaseSuggestions),
		},
		Purchase: purchase,
	}
}

// BuildInventoryPurchaseAddAnotherPage creates the Inventory Purchases
// "add another?" interstitial.
func BuildInventoryPurchaseAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, purchase domain.InventoryPurchase, cashFlowComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Inventory Purchases | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         purchase.Category,
		Answers:          inventoryPurchaseAnsweredFields(purchase, ""),
		ItemNounSingular: "inventory purchase",
		DoneAction:       fmt.Sprintf("/plan/%s/cash-flow/inventory-purchases/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/cash-flow/inventory-purchases/new", plan.ID()),
	}
}

// distributionListItem formats a Distribution for the shared section-list
// page.
func distributionListItem(planID uuid.UUID, item DistributionItem) SectionListItem {
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Distribution.Name,
		Detail:       formatMoney(item.Distribution.MonthlyAmount) + "/mo",
		EditURL:      CashFlowSectionStepURL(planID, item.ID, domain.SectionDistributions, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/cash-flow/distributions/%s", planID, item.ID),
	}
}

// BuildDistributionListPage creates the Distributions section's list page.
func BuildDistributionListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []DistributionItem, complete bool, draftItemID *uuid.UUID, draftStep string, cashFlowComplete bool) SectionListPage {
	base := BuildBasePage(r, "Distributions | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = distributionListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = CashFlowSectionStepURL(plan.ID(), *draftItemID, domain.SectionDistributions, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Cash Flow",
		SectionIcon:        "bi-cash-stack",
		SectionTitle:       "Distributions",
		SectionDescription: "Owner draws, dividends, and loan/credit repayments.",
		EmptyIcon:          "bi-wallet2",
		EmptyText:          "No distributions planned.",
		ItemNounSingular:   "distribution",
		AddButtonLabel:     "Add Outflow",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/cash-flow/distributions/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/cash-flow/distributions/finish", plan.ID()),
		OverviewURL:        CashFlowSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildDistributionStepPage creates a DistributionStepPage.
func BuildDistributionStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, dist domain.Distribution, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, cashFlowComplete bool) DistributionStepPage {
	meta := cashFlowSectionMetaByKey(domain.SectionDistributions)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete
	return DistributionStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      CashFlowSectionStepURL(plan.ID(), itemID, domain.SectionDistributions, step),
			BackURL:         backURL,
			CancelURL:       CashFlowSectionListURL(plan.ID(), domain.SectionDistributions),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: distributionAnsweredFields(dist, step),
			Suggestions:     suggestionsForStep(step, "name", distributionSuggestions),
		},
		Distribution: dist,
	}
}

// BuildDistributionAddAnotherPage creates the Distributions "add another?"
// interstitial.
func BuildDistributionAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, dist domain.Distribution, cashFlowComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Distributions | Business Planning Tool", user)
	base.CashFlowComplete = cashFlowComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         dist.Name,
		Answers:          distributionAnsweredFields(dist, ""),
		ItemNounSingular: "distribution",
		DoneAction:       fmt.Sprintf("/plan/%s/cash-flow/distributions/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/cash-flow/distributions/new", plan.ID()),
	}
}

// --- Renderers ---

// RenderInventoryPurchaseStepPage renders a single Inventory Purchase
// wizard step.
func RenderInventoryPurchaseStepPage(w http.ResponseWriter, tc map[string]*template.Template, page InventoryPurchaseStepPage) {
	renderTemplate(w, tc, "cash-flow-inventory-step.html", "base", page)
}

// RenderInventoryPurchaseStepPageWithStatus renders an Inventory Purchase
// wizard step with a custom status code (used on validation errors).
func RenderInventoryPurchaseStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page InventoryPurchaseStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "cash-flow-inventory-step.html", "base", page, statusCode)
}

// RenderDistributionStepPage renders a single Distribution wizard step.
func RenderDistributionStepPage(w http.ResponseWriter, tc map[string]*template.Template, page DistributionStepPage) {
	renderTemplate(w, tc, "cash-flow-distributions-step.html", "base", page)
}

// RenderDistributionStepPageWithStatus renders a Distribution wizard step
// with a custom status code (used on validation errors).
func RenderDistributionStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page DistributionStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "cash-flow-distributions-step.html", "base", page, statusCode)
}
