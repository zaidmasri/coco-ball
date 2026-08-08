package views

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// --- Section metadata ---
//
// sectionMeta (shared type, see wizard_shared.go) is the single source of
// truth for a section's display copy. Both the summary table and the
// section intro/landing page read from this, so the two can never drift
// apart. Starting Point's own instance is startingPointSectionMetas below;
// Payroll/Sales Forecast/Cash Flow each have their own in their own files.

var startingPointSectionMetas = []sectionMeta{
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

func startingPointSectionMetaByKey(key string) sectionMeta {
	return sectionMetaByKey(startingPointSectionMetas, key)
}

// IsStartingPointComplete reports whether every Starting Point sub-section
// is complete, given the section-status map the store returns. Used to
// decide whether the sidebar's Starting Point nav icon renders filled.
func IsStartingPointComplete(sectionStatus map[string]bool) bool {
	return hubComplete(sectionStatus, startingPointSectionMetas)
}

// --- Answered-so-far summaries ---

// AnsweredField is one row in the "answered so far" summary box shown above
// a Starting Point wizard question, and the full-answers summary shown on
// each repeatable section's "Add another?" interstitial.
type AnsweredField struct {
	Label string
	Value string
}

// depreciationMethodLabel humanizes the raw stored enum ("StraightLine")
// for read-only display ("Straight Line"). Duplicates the <option> label
// text hardcoded in starting-point-fixed-assets-step.html's <select> -
// update both if the wording ever changes.
func depreciationMethodLabel(m domain.DepreciationMethod) string {
	switch m {
	case domain.StraightLine:
		return "Straight Line"
	case domain.DoubleDeclining:
		return "Double Declining"
	case domain.None:
		return "None"
	default:
		return string(m)
	}
}

var fixedAssetFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.CapitalAsset) string
}{
	{"name", "Asset Name", func(a domain.CapitalAsset) string { return a.Name }},
	{"cost", "Cost", func(a domain.CapitalAsset) string { return formatMoney(a.PurchaseCost) }},
	{"depreciation-method", "Depreciation", func(a domain.CapitalAsset) string { return depreciationMethodLabel(a.DepreciationMethod) }},
	{"useful-life", "Useful Life", func(a domain.CapitalAsset) string { return fmt.Sprintf("%d years", a.UsefulLifeYears()) }},
}

// fixedAssetAnsweredFields returns the fields answered strictly before
// uptoStep, in wizard order. uptoStep == "" returns every field (used by
// the "Add another?" page). "useful-life" is always omitted when
// DepreciationMethod is None, since that question is never asked in that
// case (mirrors the finishNow skip in handlers.PostFixedAssetStep).
func fixedAssetAnsweredFields(asset domain.CapitalAsset, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range fixedAssetFieldDefs {
		if f.Step == uptoStep {
			break
		}
		if f.Step == "useful-life" && asset.DepreciationMethod == domain.None {
			continue
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(asset)})
	}
	return out
}

var startupCostFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.StartupCost) string
}{
	{"name", "Name", func(c domain.StartupCost) string { return c.Name }},
	{"amount", "Amount", func(c domain.StartupCost) string { return formatMoney(c.Amount) }},
}

// startupCostAnsweredFields is the Startup Costs equivalent of
// fixedAssetAnsweredFields.
func startupCostAnsweredFields(cost domain.StartupCost, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range startupCostFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(cost)})
	}
	return out
}

var fundingSourceFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.FundingSource) string
}{
	{"name", "Name", func(f domain.FundingSource) string { return f.Name }},
	{"amount", "Amount", func(f domain.FundingSource) string { return formatMoney(f.Amount) }},
	{"interest-rate", "Interest Rate", func(f domain.FundingSource) string { return formatPercent(f.InterestRatePercent()) }},
	{"term", "Term", func(f domain.FundingSource) string { return fmt.Sprintf("%d months", f.TermMonths) }},
}

// fundingSourceAnsweredFields is the Funding Sources equivalent of
// fixedAssetAnsweredFields.
func fundingSourceAnsweredFields(funding domain.FundingSource, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range fundingSourceFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(funding)})
	}
	return out
}

var cashOnHandFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.StartingBalances) string
}{
	{"cash", "Cash", func(b domain.StartingBalances) string { return formatMoney(b.Cash) }},
	{"accounts-receivable", "Accounts Receivable", func(b domain.StartingBalances) string { return formatMoney(b.AccountsReceivable) }},
	{"prepaid-expenses", "Prepaid Expenses", func(b domain.StartingBalances) string { return formatMoney(b.PrepaidExpenses) }},
	{"accounts-payable", "Accounts Payable", func(b domain.StartingBalances) string { return formatMoney(b.AccountsPayable) }},
	{"accrued-expenses", "Accrued Expenses", func(b domain.StartingBalances) string { return formatMoney(b.AccruedExpenses) }},
}

// cashOnHandAnsweredFields is the Cash on Hand equivalent of
// fixedAssetAnsweredFields.
func cashOnHandAnsweredFields(b domain.StartingBalances, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range cashOnHandFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(b)})
	}
	return out
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

// QuestionStepPage carries the fields common to every Starting Point
// wizard question, regardless of section. The shared "question-step"
// component template renders purely off these fields; each section-
// specific step page type below just adds whichever answer-so-far value
// it carries (Asset/Cost/Funding/Balances).
type QuestionStepPage struct {
	BasePage
	Plan            *domain.Plan
	SectionTitle    string
	SectionIcon     string
	Step            string
	StepNumber      int
	TotalSteps      int
	FormAction      string
	BackURL         string // empty on the first step
	CancelURL       string // used instead of BackURL when there is no previous step
	ButtonLabel     string // "Next" or "Finish"
	ErrorMessage    string
	PreviousAnswers []AnsweredField // fields answered on earlier steps of this item
	Suggestions     []string        // quick-fill pills shown under the primary text field, if any
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

// --- Suggestion pills ---

var (
	fixedAssetSuggestions    = []string{"Real Estate-Land", "Real Estate-Buildings", "Leasehold Improvements", "Equipment", "Furniture and Fixtures", "Vehicles"}
	startupCostSuggestions   = []string{"Pre-Opening Salaries and Wages", "Prepaid Insurance Premiums", "Inventory", "Legal and Accounting", "Rent Deposits", "Utility Deposits", "Supplies", "Advertising and Promotions", "Licenses", "Other Initial Startup Costs"}
	fundingSourceSuggestions = []string{"Owners Equity", "Outside Investors", "Commercial Loan", "Commercial Mortgage", "Credit Card Debt", "Vehicle Loans", "Other Bank Debt"}
)

// suggestionsForStep returns suggestions when step is the one step they
// belong to (always the section's primary text-entry step), nil otherwise.
func suggestionsForStep(step, targetStep string, suggestions []string) []string {
	if step == targetStep {
		return suggestions
	}
	return nil
}

// --- Builders ---

// BuildStartingPointSummaryPage creates the Starting Point overview page.
// sectionStatus and the item counts are computed by the store layer (the
// handler queries them), following the same pattern as BuildSetupPage
// taking pre-computed invites/isOwner.
func BuildStartingPointSummaryPage(r *http.Request, user *domain.User, plan *domain.Plan, sectionStatus map[string]bool, fixedAssetCount, startupCostCount, fundingSourceCount int) HubSummaryPage {
	base := BuildBasePage(r, "Starting Point | Business Planning Tool", user)
	base.StartingPointComplete = IsStartingPointComplete(sectionStatus)

	counts := map[string]int{
		domain.SectionFixedAssets:    fixedAssetCount,
		domain.SectionStartupCosts:   startupCostCount,
		domain.SectionFundingSources: fundingSourceCount,
	}

	sections := make([]HubSectionStatus, len(startingPointSectionMetas))
	for i, m := range startingPointSectionMetas {
		sections[i] = HubSectionStatus{
			Key:           m.Key,
			Title:         m.Title,
			Description:   m.ShortDesc,
			Complete:      sectionStatus[m.Key],
			ItemCount:     counts[m.Key],
			EditURL:       SectionListURL(plan.ID(), m.Key),
			GetStartedURL: fmt.Sprintf("/plan/%s/starting-point/intro/%s", plan.ID(), m.Key),
		}
	}

	return HubSummaryPage{
		BasePage:       base,
		Plan:           plan,
		StepBadge:      "Step 2 of 9",
		HubTitle:       "Starting Point",
		HubDescription: "Let's establish your baseline, one section at a time. Work through each section below — you can leave and come back whenever you like.",
		Sections:       sections,
		BackURL:        fmt.Sprintf("/plan/%s/setup", plan.ID()),
		ContinueURL:    fmt.Sprintf("/plan/%s/payroll", plan.ID()),
	}
}

// BuildStartingPointSectionIntroPage creates the landing/description page
// shown when a user clicks "Get Started" on a section. section must be one
// of the domain.Section* constants.
func BuildStartingPointSectionIntroPage(r *http.Request, user *domain.User, plan *domain.Plan, section string, startingPointComplete bool) SectionIntroPage {
	meta := startingPointSectionMetaByKey(section)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete

	// SectionListURL is the right "continue" target for every section: for
	// the 3 repeatable sections it's their list/entry page, and for Cash
	// on Hand it happens to be the exact same URL shape as its own
	// entry-redirect route (GetCashOnHandEntry), which resumes/starts the
	// singleton wizard automatically.
	continueURL := SectionListURL(plan.ID(), section)

	return SectionIntroPage{
		BasePage:      base,
		Plan:          plan,
		HubBadge:      "Starting Point",
		HubSummaryURL: StartingPointSummaryURL(plan.ID()),
		SectionTitle:  meta.Title,
		SectionIcon:   meta.Icon,
		Description:   meta.LongDesc,
		ContinueURL:   continueURL,
	}
}

// fixedAssetListItem formats a Fixed Asset for the shared section-list page.
func fixedAssetListItem(planID uuid.UUID, item StartingPointFixedAssetItem) SectionListItem {
	detail := formatMoney(item.Asset.PurchaseCost)
	if item.Asset.DepreciationMethod == domain.None {
		detail += " · No depreciation"
	} else {
		detail += fmt.Sprintf(" · %s, %d yrs", item.Asset.DepreciationMethod, item.Asset.UsefulLifeYears())
	}
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Asset.Name,
		Detail:       detail,
		EditURL:      SectionStepURL(planID, item.ID, domain.SectionFixedAssets, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/starting-point/fixed-assets/%s", planID, item.ID),
	}
}

// BuildFixedAssetsListPage creates the Fixed Assets section's list page.
func BuildFixedAssetsListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointFixedAssetItem, complete bool, draftItemID *uuid.UUID, draftStep string, startingPointComplete bool) SectionListPage {
	base := BuildBasePage(r, "Fixed Assets | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = fixedAssetListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = SectionStepURL(plan.ID(), *draftItemID, domain.SectionFixedAssets, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Starting Point",
		SectionIcon:        "bi-building",
		SectionTitle:       "Fixed Assets",
		SectionDescription: "Equipment, vehicles, real estate, and other assets you'll purchase.",
		EmptyIcon:          "bi-building-add",
		EmptyText:          "No fixed assets yet.",
		ItemNounSingular:   "fixed asset",
		AddButtonLabel:     "Add Asset",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/starting-point/fixed-assets/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/starting-point/fixed-assets/finish", plan.ID()),
		OverviewURL:        StartingPointSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// startupCostListItem formats a Startup Cost for the shared section-list page.
func startupCostListItem(planID uuid.UUID, item StartingPointStartupCostItem) SectionListItem {
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Cost.Name,
		Detail:       formatMoney(item.Cost.Amount),
		EditURL:      SectionStepURL(planID, item.ID, domain.SectionStartupCosts, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/starting-point/startup-costs/%s", planID, item.ID),
	}
}

// BuildStartupCostsListPage creates the Startup Costs section's list page.
func BuildStartupCostsListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointStartupCostItem, complete bool, draftItemID *uuid.UUID, draftStep string, startingPointComplete bool) SectionListPage {
	base := BuildBasePage(r, "Startup Costs | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = startupCostListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = SectionStepURL(plan.ID(), *draftItemID, domain.SectionStartupCosts, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Starting Point",
		SectionIcon:        "bi-cash-coin",
		SectionTitle:       "Startup Costs",
		SectionDescription: "One-time costs to get the business up and running.",
		EmptyIcon:          "bi-cash-stack",
		EmptyText:          "No startup costs yet.",
		ItemNounSingular:   "startup cost",
		AddButtonLabel:     "Add Startup Cost",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/starting-point/startup-costs/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/starting-point/startup-costs/finish", plan.ID()),
		OverviewURL:        StartingPointSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// fundingSourceListItem formats a Funding Source for the shared
// section-list page.
func fundingSourceListItem(planID uuid.UUID, item StartingPointFundingSourceItem) SectionListItem {
	detail := formatMoney(item.Funding.Amount)
	if item.Funding.InterestRate != 0 || item.Funding.TermMonths != 0 {
		detail += fmt.Sprintf(" · %s over %d mo", formatPercent(item.Funding.InterestRatePercent()), item.Funding.TermMonths)
	} else {
		detail += " · Equity (no loan)"
	}
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Funding.Name,
		Detail:       detail,
		EditURL:      SectionStepURL(planID, item.ID, domain.SectionFundingSources, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/starting-point/funding-sources/%s", planID, item.ID),
	}
}

// BuildFundingSourcesListPage creates the Funding Sources section's list page.
func BuildFundingSourcesListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []StartingPointFundingSourceItem, complete bool, draftItemID *uuid.UUID, draftStep string, startingPointComplete bool) SectionListPage {
	base := BuildBasePage(r, "Funding Sources | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = fundingSourceListItem(plan.ID(), item)
	}

	var draftStepURL string
	if draftItemID != nil {
		draftStepURL = SectionStepURL(plan.ID(), *draftItemID, domain.SectionFundingSources, draftStep)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Starting Point",
		SectionIcon:        "bi-piggy-bank",
		SectionTitle:       "Funding Sources",
		SectionDescription: "Loans and owner investment used to finance the business.",
		EmptyIcon:          "bi-bank",
		EmptyText:          "No funding sources yet.",
		ItemNounSingular:   "funding source",
		AddButtonLabel:     "Add Funding Source",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		AddAction:          fmt.Sprintf("/plan/%s/starting-point/funding-sources/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/starting-point/funding-sources/finish", plan.ID()),
		OverviewURL:        StartingPointSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildFixedAssetStepPage creates a StartingPointFixedAssetStepPage.
func BuildFixedAssetStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, asset domain.CapitalAsset, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, startingPointComplete bool) StartingPointFixedAssetStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionFixedAssets)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return StartingPointFixedAssetStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SectionStepURL(plan.ID(), itemID, domain.SectionFixedAssets, step),
			BackURL:         backURL,
			CancelURL:       SectionListURL(plan.ID(), domain.SectionFixedAssets),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: fixedAssetAnsweredFields(asset, step),
			Suggestions:     suggestionsForStep(step, "name", fixedAssetSuggestions),
		},
		Asset: asset,
	}
}

// BuildStartupCostStepPage creates a StartingPointStartupCostStepPage.
func BuildStartupCostStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, cost domain.StartupCost, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, startingPointComplete bool) StartingPointStartupCostStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionStartupCosts)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return StartingPointStartupCostStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SectionStepURL(plan.ID(), itemID, domain.SectionStartupCosts, step),
			BackURL:         backURL,
			CancelURL:       SectionListURL(plan.ID(), domain.SectionStartupCosts),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: startupCostAnsweredFields(cost, step),
			Suggestions:     suggestionsForStep(step, "name", startupCostSuggestions),
		},
		Cost: cost,
	}
}

// BuildFundingSourceStepPage creates a StartingPointFundingSourceStepPage.
func BuildFundingSourceStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, funding domain.FundingSource, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, startingPointComplete bool) StartingPointFundingSourceStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionFundingSources)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return StartingPointFundingSourceStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SectionStepURL(plan.ID(), itemID, domain.SectionFundingSources, step),
			BackURL:         backURL,
			CancelURL:       SectionListURL(plan.ID(), domain.SectionFundingSources),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: fundingSourceAnsweredFields(funding, step),
			Suggestions:     suggestionsForStep(step, "name", fundingSourceSuggestions),
		},
		Funding: funding,
	}
}

// BuildCashOnHandStepPage creates a StartingPointCashOnHandStepPage.
func BuildCashOnHandStepPage(r *http.Request, user *domain.User, plan *domain.Plan, balances domain.StartingBalances, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, startingPointComplete bool) StartingPointCashOnHandStepPage {
	meta := startingPointSectionMetaByKey(domain.SectionCashOnHand)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return StartingPointCashOnHandStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SectionSingletonStepURL(plan.ID(), domain.SectionCashOnHand, step),
			BackURL:         backURL,
			CancelURL:       StartingPointSummaryURL(plan.ID()), // Cash on Hand has no list page to cancel back to
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: cashOnHandAnsweredFields(balances, step),
		},
		Balances: balances,
	}
}

// BuildFixedAssetsAddAnotherPage creates the Fixed Assets "add another?"
// interstitial.
func BuildFixedAssetsAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, asset domain.CapitalAsset, startingPointComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Fixed Assets | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         asset.Name,
		Answers:          fixedAssetAnsweredFields(asset, ""),
		ItemNounSingular: "fixed asset",
		DoneAction:       fmt.Sprintf("/plan/%s/starting-point/fixed-assets/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/starting-point/fixed-assets/new", plan.ID()),
	}
}

// BuildStartupCostsAddAnotherPage creates the Startup Costs "add another?"
// interstitial.
func BuildStartupCostsAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, cost domain.StartupCost, startingPointComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Startup Costs | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         cost.Name,
		Answers:          startupCostAnsweredFields(cost, ""),
		ItemNounSingular: "startup cost",
		DoneAction:       fmt.Sprintf("/plan/%s/starting-point/startup-costs/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/starting-point/startup-costs/new", plan.ID()),
	}
}

// BuildFundingSourcesAddAnotherPage creates the Funding Sources "add
// another?" interstitial.
func BuildFundingSourcesAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, funding domain.FundingSource, startingPointComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Funding Sources | Business Planning Tool", user)
	base.StartingPointComplete = startingPointComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         funding.Name,
		Answers:          fundingSourceAnsweredFields(funding, ""),
		ItemNounSingular: "funding source",
		DoneAction:       fmt.Sprintf("/plan/%s/starting-point/funding-sources/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/starting-point/funding-sources/new", plan.ID()),
	}
}

// --- Renderers ---

// RenderFixedAssetStepPage renders a single Fixed Assets wizard step
func RenderFixedAssetStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetStepPage) {
	renderTemplate(w, tc, "starting-point-fixed-assets-step.html", "base", page)
}

// RenderFixedAssetStepPageWithStatus renders a Fixed Assets wizard step
// with a custom status code (used on validation errors)
func RenderFixedAssetStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFixedAssetStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-fixed-assets-step.html", "base", page, statusCode)
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

// RenderFundingSourceStepPage renders a single Funding Sources wizard step
func RenderFundingSourceStepPage(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourceStepPage) {
	renderTemplate(w, tc, "starting-point-funding-sources-step.html", "base", page)
}

// RenderFundingSourceStepPageWithStatus renders a Funding Sources wizard
// step with a custom status code (used on validation errors)
func RenderFundingSourceStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page StartingPointFundingSourceStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "starting-point-funding-sources-step.html", "base", page, statusCode)
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
