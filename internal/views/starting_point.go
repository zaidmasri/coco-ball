package views

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// --- Section metadata ---

// startingPointSectionMeta is the single source of truth for a Starting
// Point section's display copy. Both the summary table and the section
// intro/landing page read from this, so the two can never drift apart.
type startingPointSectionMeta struct {
	Key       string
	Title     string
	Icon      string // bootstrap-icons class, e.g. "bi-building"
	ShortDesc string // one-liner used on the summary table and list page header
	LongDesc  string // fuller paragraph used on the intro/landing page
}

var startingPointSectionMetas = []startingPointSectionMeta{
	{
		Key:       domain.SectionFixedAssets,
		Title:     "Fixed Assets",
		Icon:      "bi-building",
		ShortDesc: "Equipment, vehicles, real estate, and other assets you'll purchase.",
		LongDesc: "Fixed assets are the equipment, vehicles, real estate, and other " +
			"long-lived purchases your business needs to operate. We'll walk through " +
			"them one at a time - what each one is, what it costs, and how it " +
			"depreciates - so your projections account for both the upfront cost and " +
			"the ongoing depreciation expense.",
	},
	{
		Key:       domain.SectionStartupCosts,
		Title:     "Startup Costs",
		Icon:      "bi-cash-coin",
		ShortDesc: "One-time costs to get the business up and running.",
		LongDesc: "Startup costs are the one-time expenses you'll pay before or " +
			"while opening - things like legal fees, deposits, initial inventory, " +
			"or licenses. These are different from your ongoing operating expenses; " +
			"we'll capture them separately so your starting balance sheet reflects " +
			"exactly what it took to get the doors open.",
	},
	{
		Key:       domain.SectionFundingSources,
		Title:     "Funding Sources",
		Icon:      "bi-piggy-bank",
		ShortDesc: "Loans and owner investment used to finance the business.",
		LongDesc: "Funding sources are how you'll pay for it all - loans, lines of " +
			"credit, or money you or other owners are putting in directly. For " +
			"anything with an interest rate and repayment term, we'll use those to " +
			"build a loan amortization schedule automatically.",
	},
	{
		Key:       domain.SectionCashOnHand,
		Title:     "Cash on Hand",
		Icon:      "bi-wallet2",
		ShortDesc: "Your starting cash, receivables, payables, and other balances.",
		LongDesc: "If you're already operating, Cash on Hand captures your existing " +
			"balance sheet - cash in the bank, money owed to you, and money you owe - " +
			"so your projections start from where the business actually stands today, " +
			"not from zero. Starting a brand-new business? Leave everything at $0.",
	},
}

func startingPointSectionMetaByKey(key string) startingPointSectionMeta {
	for _, m := range startingPointSectionMetas {
		if m.Key == key {
			return m
		}
	}
	return startingPointSectionMeta{Key: key, Title: key}
}

// --- URL helpers ---
//
// Exported so both this package's builders and internal/handlers can build
// Starting Point URLs from the same single implementation, rather than
// each maintaining its own copy of the same path patterns.

// StartingPointSummaryURL is the Starting Point overview page for a plan.
func StartingPointSummaryURL(planID uuid.UUID) string {
	return fmt.Sprintf("/plan/%s/starting-point", planID)
}

// SectionListURL is a repeatable section's list/entry page.
func SectionListURL(planID uuid.UUID, section string) string {
	return fmt.Sprintf("/plan/%s/starting-point/%s", planID, section)
}

// SectionStepURL is one wizard question for a specific item in a
// repeatable section (Fixed Assets, Startup Costs, Funding Sources).
func SectionStepURL(planID, itemID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/starting-point/%s/%s/%s", planID, section, itemID, step)
}

// SectionSingletonStepURL is one wizard question for a singleton section
// (Cash on Hand), which has no item ID.
func SectionSingletonStepURL(planID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/starting-point/%s/%s", planID, section, step)
}

// --- Page types ---

// StartingPointSummaryPage is the data for the Starting Point overview
// page: one row per section, each with its own completion status and a
// Get Started/Edit link into that section's wizard.
type StartingPointSummaryPage struct {
	BasePage
	Plan     *domain.Plan
	Sections []StartingPointSectionStatus
}

// StartingPointSectionStatus is one row of the Starting Point summary
// table, rendered by the "sections-table" component.
type StartingPointSectionStatus struct {
	Key         string // route segment, e.g. domain.SectionFixedAssets
	Title       string
	Description string
	Complete    bool
	ItemCount   int // unused (0) for Cash on Hand, which isn't repeatable
}

// StartingPointSectionIntroPage is a section's landing/description page,
// shown when the user clicks "Get Started" on a not-yet-started section.
//
// SectionTitle/SectionIcon (not Title/Icon) deliberately, matching
// QuestionStepPage's naming: a plain "Title" field here would shadow the
// embedded BasePage.Title that base.html's <title> tag actually needs
// (Go's field-resolution rule prefers the shallower field), silently
// dropping the "| Business Planning Tool" suffix from the browser tab.
type StartingPointSectionIntroPage struct {
	BasePage
	Plan         *domain.Plan
	Section      string
	SectionTitle string
	SectionIcon  string
	Description  string
	ContinueURL  string
}

// QuestionStepPage carries the fields common to every Starting Point
// wizard question, regardless of section. The shared "question-step"
// component template renders purely off these fields; each section-
// specific step page type below just adds whichever answer-so-far value
// it carries (Asset/Cost/Funding/Balances).
type QuestionStepPage struct {
	BasePage
	Plan         *domain.Plan
	SectionTitle string
	SectionIcon  string
	Step         string
	StepNumber   int
	TotalSteps   int
	FormAction   string
	BackURL      string // empty on the first step
	CancelURL    string // used instead of BackURL when there is no previous step
	ButtonLabel  string // "Next" or "Finish"
	ErrorMessage string
}

// StartingPointFixedAssetItem pairs a Fixed Asset's own wizard-item ID
// with its underlying domain value, so the list page can render
// Edit/Delete links (which need the item ID) without CapitalAsset itself
// needing to know about wizard/storage identity.
type StartingPointFixedAssetItem struct {
	ID    uuid.UUID
	Asset domain.CapitalAsset
}

// StartingPointStartupCostItem is the Startup Costs equivalent of
// StartingPointFixedAssetItem.
type StartingPointStartupCostItem struct {
	ID   uuid.UUID
	Cost domain.StartupCost
}

// StartingPointFundingSourceItem is the Funding Sources equivalent of
// StartingPointFixedAssetItem.
type StartingPointFundingSourceItem struct {
	ID      uuid.UUID
	Funding domain.FundingSource
}

// StartingPointFixedAssetsListPage is the Fixed Assets section's list
// page: every finished asset, plus a Resume link if the user left an item
// mid-wizard.
type StartingPointFixedAssetsListPage struct {
	BasePage
	Plan        *domain.Plan
	Items       []StartingPointFixedAssetItem
	Complete    bool
	DraftItemID *uuid.UUID
	DraftStep   string
}

// StartingPointStartupCostsListPage is the Startup Costs equivalent of
// StartingPointFixedAssetsListPage.
type StartingPointStartupCostsListPage struct {
	BasePage
	Plan        *domain.Plan
	Items       []StartingPointStartupCostItem
	Complete    bool
	DraftItemID *uuid.UUID
	DraftStep   string
}

// StartingPointFundingSourcesListPage is the Funding Sources equivalent of
// StartingPointFixedAssetsListPage.
type StartingPointFundingSourcesListPage struct {
	BasePage
	Plan        *domain.Plan
	Items       []StartingPointFundingSourceItem
	Complete    bool
	DraftItemID *uuid.UUID
	DraftStep   string
}

// StartingPointFixedAssetStepPage renders a single wizard question for one
// Fixed Asset item.
type StartingPointFixedAssetStepPage struct {
	QuestionStepPage
	Asset domain.CapitalAsset // answers so far, for pre-fill on Back/error
}

// StartingPointStartupCostStepPage is the Startup Costs equivalent of
// StartingPointFixedAssetStepPage.
type StartingPointStartupCostStepPage struct {
	QuestionStepPage
	Cost domain.StartupCost
}

// StartingPointFundingSourceStepPage is the Funding Sources equivalent of
// StartingPointFixedAssetStepPage.
type StartingPointFundingSourceStepPage struct {
	QuestionStepPage
	Funding domain.FundingSource
}

// StartingPointCashOnHandStepPage renders a single wizard question for the
// plan's Cash on Hand singleton.
type StartingPointCashOnHandStepPage struct {
	QuestionStepPage
	Balances domain.StartingBalances
}

// StartingPointFixedAssetsAddAnotherPage is the "Add another?" interstitial
// shown after finishing one Fixed Asset item.
type StartingPointFixedAssetsAddAnotherPage struct {
	BasePage
	Plan  *domain.Plan
	Asset domain.CapitalAsset // the item just finished, for a confirmation summary
}

// StartingPointStartupCostsAddAnotherPage is the Startup Costs equivalent
// of StartingPointFixedAssetsAddAnotherPage.
type StartingPointStartupCostsAddAnotherPage struct {
	BasePage
	Plan *domain.Plan
	Cost domain.StartupCost
}

// StartingPointFundingSourcesAddAnotherPage is the Funding Sources
// equivalent of StartingPointFixedAssetsAddAnotherPage.
type StartingPointFundingSourcesAddAnotherPage struct {
	BasePage
	Plan    *domain.Plan
	Funding domain.FundingSource
}

// --- Builders ---

// BuildStartingPointSummaryPage creates the Starting Point overview page.
// sectionStatus and the item counts are computed by the store layer (the
// handler queries them), following the same pattern as BuildSetupPage
// taking pre-computed invites/isOwner.
func BuildStartingPointSummaryPage(r *http.Request, user *domain.User, plan *domain.Plan, sectionStatus map[string]bool, fixedAssetCount, startupCostCount, fundingSourceCount int) StartingPointSummaryPage {
	base := BuildBasePage(r, "Starting Point | Business Planning Tool", user)

	counts := map[string]int{
		domain.SectionFixedAssets:    fixedAssetCount,
		domain.SectionStartupCosts:   startupCostCount,
		domain.SectionFundingSources: fundingSourceCount,
	}

	sections := make([]StartingPointSectionStatus, len(startingPointSectionMetas))
	for i, m := range startingPointSectionMetas {
		sections[i] = StartingPointSectionStatus{
			Key:         m.Key,
			Title:       m.Title,
			Description: m.ShortDesc,
			Complete:    sectionStatus[m.Key],
			ItemCount:   counts[m.Key],
		}
	}

	return StartingPointSummaryPage{
		BasePage: base,
		Plan:     plan,
		Sections: sections,
	}
}

// BuildStartingPointSectionIntroPage creates the landing/description page
// shown when a user clicks "Get Started" on a section. section must be one
// of the domain.Section* constants.
func BuildStartingPointSectionIntroPage(r *http.Request, user *domain.User, plan *domain.Plan, section string) StartingPointSectionIntroPage {
	meta := startingPointSectionMetaByKey(section)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)

	// SectionListURL is the right "continue" target for every section: for
	// the 3 repeatable sections it's their list/entry page, and for Cash
	// on Hand it happens to be the exact same URL shape as its own
	// entry-redirect route (GetCashOnHandEntry), which resumes/starts the
	// singleton wizard automatically.
	continueURL := SectionListURL(plan.ID(), section)

	return StartingPointSectionIntroPage{
		BasePage:     base,
		Plan:         plan,
		Section:      section,
		SectionTitle: meta.Title,
		SectionIcon:  meta.Icon,
		Description:  meta.LongDesc,
		ContinueURL:  continueURL,
	}
}

// BuildFixedAssetsListPage creates a StartingPointFixedAssetsListPage.
func BuildFixedAssetsListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointFixedAssetItem, complete bool, draftItemID *uuid.UUID, draftStep string) StartingPointFixedAssetsListPage {
	base := BuildBasePage(r, "Fixed Assets | Business Planning Tool", user)
	return StartingPointFixedAssetsListPage{
		BasePage:    base,
		Plan:        plan,
		Items:       items,
		Complete:    complete,
		DraftItemID: draftItemID,
		DraftStep:   draftStep,
	}
}

// BuildStartupCostsListPage creates a StartingPointStartupCostsListPage.
func BuildStartupCostsListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointStartupCostItem, complete bool, draftItemID *uuid.UUID, draftStep string) StartingPointStartupCostsListPage {
	base := BuildBasePage(r, "Startup Costs | Business Planning Tool", user)
	return StartingPointStartupCostsListPage{
		BasePage:    base,
		Plan:        plan,
		Items:       items,
		Complete:    complete,
		DraftItemID: draftItemID,
		DraftStep:   draftStep,
	}
}

// BuildFundingSourcesListPage creates a StartingPointFundingSourcesListPage.
func BuildFundingSourcesListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointFundingSourceItem, complete bool, draftItemID *uuid.UUID, draftStep string) StartingPointFundingSourcesListPage {
	base := BuildBasePage(r, "Funding Sources | Business Planning Tool", user)
	return StartingPointFundingSourcesListPage{
		BasePage:    base,
		Plan:        plan,
		Items:       items,
		Complete:    complete,
		DraftItemID: draftItemID,
		DraftStep:   draftStep,
	}
}

// BuildFixedAssetStepPage creates a StartingPointFixedAssetStepPage.
func BuildFixedAssetStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, asset domain.CapitalAsset, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string) StartingPointFixedAssetStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionFixedAssets)
	return StartingPointFixedAssetStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:     BuildBasePage(r, meta.Title+" | Business Planning Tool", user),
			Plan:         plan,
			SectionTitle: meta.Title,
			SectionIcon:  meta.Icon,
			Step:         step,
			StepNumber:   stepNumber,
			TotalSteps:   totalSteps,
			FormAction:   SectionStepURL(plan.ID(), itemID, domain.SectionFixedAssets, step),
			BackURL:      backURL,
			CancelURL:    SectionListURL(plan.ID(), domain.SectionFixedAssets),
			ButtonLabel:  buttonLabel,
			ErrorMessage: errMsg,
		},
		Asset: asset,
	}
}

// BuildStartupCostStepPage creates a StartingPointStartupCostStepPage.
func BuildStartupCostStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, cost domain.StartupCost, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string) StartingPointStartupCostStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionStartupCosts)
	return StartingPointStartupCostStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:     BuildBasePage(r, meta.Title+" | Business Planning Tool", user),
			Plan:         plan,
			SectionTitle: meta.Title,
			SectionIcon:  meta.Icon,
			Step:         step,
			StepNumber:   stepNumber,
			TotalSteps:   totalSteps,
			FormAction:   SectionStepURL(plan.ID(), itemID, domain.SectionStartupCosts, step),
			BackURL:      backURL,
			CancelURL:    SectionListURL(plan.ID(), domain.SectionStartupCosts),
			ButtonLabel:  buttonLabel,
			ErrorMessage: errMsg,
		},
		Cost: cost,
	}
}

// BuildFundingSourceStepPage creates a StartingPointFundingSourceStepPage.
func BuildFundingSourceStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, funding domain.FundingSource, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string) StartingPointFundingSourceStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionFundingSources)
	return StartingPointFundingSourceStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:     BuildBasePage(r, meta.Title+" | Business Planning Tool", user),
			Plan:         plan,
			SectionTitle: meta.Title,
			SectionIcon:  meta.Icon,
			Step:         step,
			StepNumber:   stepNumber,
			TotalSteps:   totalSteps,
			FormAction:   SectionStepURL(plan.ID(), itemID, domain.SectionFundingSources, step),
			BackURL:      backURL,
			CancelURL:    SectionListURL(plan.ID(), domain.SectionFundingSources),
			ButtonLabel:  buttonLabel,
			ErrorMessage: errMsg,
		},
		Funding: funding,
	}
}

// BuildCashOnHandStepPage creates a StartingPointCashOnHandStepPage.
func BuildCashOnHandStepPage(r *http.Request, user *domain.User, plan *domain.Plan, balances domain.StartingBalances, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string) StartingPointCashOnHandStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionCashOnHand)
	return StartingPointCashOnHandStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:     BuildBasePage(r, meta.Title+" | Business Planning Tool", user),
			Plan:         plan,
			SectionTitle: meta.Title,
			SectionIcon:  meta.Icon,
			Step:         step,
			StepNumber:   stepNumber,
			TotalSteps:   totalSteps,
			FormAction:   SectionSingletonStepURL(plan.ID(), domain.SectionCashOnHand, step),
			BackURL:      backURL,
			CancelURL:    StartingPointSummaryURL(plan.ID()), // Cash on Hand has no list page to cancel back to
			ButtonLabel:  buttonLabel,
			ErrorMessage: errMsg,
		},
		Balances: balances,
	}
}

// BuildFixedAssetsAddAnotherPage creates a
// StartingPointFixedAssetsAddAnotherPage.
func BuildFixedAssetsAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, asset domain.CapitalAsset) StartingPointFixedAssetsAddAnotherPage {
	base := BuildBasePage(r, "Fixed Assets | Business Planning Tool", user)
	return StartingPointFixedAssetsAddAnotherPage{
		BasePage: base,
		Plan:     plan,
		Asset:    asset,
	}
}

// BuildStartupCostsAddAnotherPage creates a
// StartingPointStartupCostsAddAnotherPage.
func BuildStartupCostsAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, cost domain.StartupCost) StartingPointStartupCostsAddAnotherPage {
	base := BuildBasePage(r, "Startup Costs | Business Planning Tool", user)
	return StartingPointStartupCostsAddAnotherPage{
		BasePage: base,
		Plan:     plan,
		Cost:     cost,
	}
}

// BuildFundingSourcesAddAnotherPage creates a
// StartingPointFundingSourcesAddAnotherPage.
func BuildFundingSourcesAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, funding domain.FundingSource) StartingPointFundingSourcesAddAnotherPage {
	base := BuildBasePage(r, "Funding Sources | Business Planning Tool", user)
	return StartingPointFundingSourcesAddAnotherPage{
		BasePage: base,
		Plan:     plan,
		Funding:  funding,
	}
}

// --- Renderers ---

// RenderStartingPointSummaryPage renders the Starting Point overview page
func RenderStartingPointSummaryPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointSummaryPage) {
	renderTemplate(w, tc, "starting-point.html", "base", page)
}

// RenderStartingPointSectionIntroPage renders a section's landing page
func RenderStartingPointSectionIntroPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointSectionIntroPage) {
	renderTemplate(w, tc, "starting-point-section-intro.html", "base", page)
}

// RenderFixedAssetsListPage renders the Fixed Assets section list page
func RenderFixedAssetsListPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetsListPage) {
	renderTemplate(w, tc, "starting-point-fixed-assets.html", "base", page)
}

// RenderFixedAssetStepPage renders a single Fixed Assets wizard step
func RenderFixedAssetStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetStepPage) {
	renderTemplate(w, tc, "starting-point-fixed-assets-step.html", "base", page)
}

// RenderFixedAssetStepPageWithStatus renders a Fixed Assets wizard step
// with a custom status code (used on validation errors)
func RenderFixedAssetStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-fixed-assets-step.html", "base", page, statusCode)
}

// RenderFixedAssetsAddAnotherPage renders the Fixed Assets "Add another?"
// interstitial
func RenderFixedAssetsAddAnotherPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetsAddAnotherPage) {
	renderTemplate(w, tc, "starting-point-fixed-assets-add-another.html", "base", page)
}

// RenderStartupCostsListPage renders the Startup Costs section list page
func RenderStartupCostsListPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointStartupCostsListPage) {
	renderTemplate(w, tc, "starting-point-startup-costs.html", "base", page)
}

// RenderStartupCostStepPage renders a single Startup Costs wizard step
func RenderStartupCostStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointStartupCostStepPage) {
	renderTemplate(w, tc, "starting-point-startup-costs-step.html", "base", page)
}

// RenderStartupCostStepPageWithStatus renders a Startup Costs wizard step
// with a custom status code (used on validation errors)
func RenderStartupCostStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointStartupCostStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-startup-costs-step.html", "base", page, statusCode)
}

// RenderStartupCostsAddAnotherPage renders the Startup Costs "Add another?"
// interstitial
func RenderStartupCostsAddAnotherPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointStartupCostsAddAnotherPage) {
	renderTemplate(w, tc, "starting-point-startup-costs-add-another.html", "base", page)
}

// RenderFundingSourcesListPage renders the Funding Sources section list page
func RenderFundingSourcesListPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourcesListPage) {
	renderTemplate(w, tc, "starting-point-funding-sources.html", "base", page)
}

// RenderFundingSourceStepPage renders a single Funding Sources wizard step
func RenderFundingSourceStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourceStepPage) {
	renderTemplate(w, tc, "starting-point-funding-sources-step.html", "base", page)
}

// RenderFundingSourceStepPageWithStatus renders a Funding Sources wizard
// step with a custom status code (used on validation errors)
func RenderFundingSourceStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourceStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-funding-sources-step.html", "base", page, statusCode)
}

// RenderFundingSourcesAddAnotherPage renders the Funding Sources
// "Add another?" interstitial
func RenderFundingSourcesAddAnotherPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourcesAddAnotherPage) {
	renderTemplate(w, tc, "starting-point-funding-sources-add-another.html", "base", page)
}

// RenderCashOnHandStepPage renders a single Cash on Hand wizard step
func RenderCashOnHandStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointCashOnHandStepPage) {
	renderTemplate(w, tc, "starting-point-cash-on-hand-step.html", "base", page)
}

// RenderCashOnHandStepPageWithStatus renders a Cash on Hand wizard step
// with a custom status code (used on validation errors)
func RenderCashOnHandStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointCashOnHandStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-cash-on-hand-step.html", "base", page, statusCode)
}
