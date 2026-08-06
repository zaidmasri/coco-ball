package views

import (
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// HubCompletion carries all four hubs' completion state at once, for pages
// that sit outside any single hub (Setup, Operating Expenses, Income
// Statement, Balance Sheet, Analytics) but whose sidebar still needs every
// icon's fill state to be accurate, not just the current page's own hub.
type HubCompletion struct {
	StartingPoint     bool
	Payroll           bool
	SalesForecast     bool
	OperatingExpenses bool
	CashFlow          bool
}

func (c HubCompletion) applyTo(base *BasePage) {
	base.StartingPointComplete = c.StartingPoint
	base.PayrollComplete = c.Payroll
	base.SalesForecastComplete = c.SalesForecast
	base.OperatingExpensesComplete = c.OperatingExpenses
	base.CashFlowComplete = c.CashFlow
}

// --- Section metadata ---

// sectionMeta is the shared per-section display-copy shape used by every
// wizard hub's own section-metadata table (e.g. startingPointSectionMetas
// in starting_point.go, payrollSectionMetas in payroll.go). The summary
// table and section intro page both read from these tables, so the two
// can never drift apart.
type sectionMeta struct {
	Key       string
	Title     string
	Icon      string // bootstrap-icons class, e.g. "bi-building"
	ShortDesc string // one-liner used on the summary table and list page header
	LongDesc  string // fuller paragraph used on the intro/landing page
}

func sectionMetaByKey(metas []sectionMeta, key string) sectionMeta {
	for _, m := range metas {
		if m.Key == key {
			return m
		}
	}
	return sectionMeta{Key: key, Title: key}
}

// hubComplete reports whether every section in metas is complete, given
// the section-status map the store returns.
func hubComplete(sectionStatus map[string]bool, metas []sectionMeta) bool {
	for _, m := range metas {
		if !sectionStatus[m.Key] {
			return false
		}
	}
	return true
}

// HubSummaryPage is the generic wizard-hub overview page: a badge/title/
// description header, the shared sections-table listing each sub-section's
// completion state, and Back/Continue footer navigation. Used by all four
// hubs (Starting Point, Payroll, Sales Forecast, Cash Flow) via the shared
// "wizard-hub-summary.html" template.
type HubSummaryPage struct {
	BasePage
	Plan           *domain.Plan
	StepBadge      string // e.g. "Step 2 of 9"
	HubTitle       string
	HubDescription string
	Sections       []HubSectionStatus
	BackURL        string
	ContinueURL    string
}

// HubSectionStatus is one row of a HubSummaryPage's sections-table.
type HubSectionStatus struct {
	Key           string
	Title         string
	Description   string
	Complete      bool
	ItemCount     int // unused (0) for singleton sections
	EditURL       string
	GetStartedURL string
}

// RenderHubSummaryPage renders any hub's overview page via the single
// shared template.
func RenderHubSummaryPage(w http.ResponseWriter, tc map[string]*template.Template, page HubSummaryPage) {
	renderTemplate(w, tc, "wizard-hub-summary.html", "base", page)
}

// SectionListPage is the generic repeatable-section list page: every
// finished item, an in-progress draft resume banner if any, and Add/Done
// actions. Used by every repeatable wizard entity across all four hubs
// (Starting Point, Payroll, Sales Forecast, Cash Flow) via the single
// shared "wizard-section-list.html" template - every display string here
// is pre-computed by each entity's own builder function, so the template
// itself has no entity-specific knowledge.
type SectionListPage struct {
	BasePage
	Plan               *domain.Plan // sidebar.html needs .Plan.ID for every nav link
	HubBadge           string
	SectionIcon        string
	SectionTitle       string
	SectionDescription string
	EmptyIcon          string
	EmptyText          string
	ItemNounSingular   string // e.g. "fixed asset" -> "an unfinished fixed asset"
	AddButtonLabel     string // e.g. "Add Asset"
	Items              []SectionListItem
	DraftItemID        *uuid.UUID
	DraftStepURL       string
	AddAction          string
	FinishAction       string
	OverviewURL        string
	Complete           bool
}

// SectionListItem is one row of a SectionListPage. Title/Detail are
// pre-formatted display strings (the same convention AnsweredField.Value
// uses), so the shared template just prints them.
type SectionListItem struct {
	ID           uuid.UUID
	Title        string
	Detail       string
	EditURL      string
	DeleteAction string
}

// AddAnotherPage is the generic "add another?" interstitial shown after
// finishing one item of a repeatable wizard section, used across all four
// hubs via the shared "wizard-add-another.html" template.
type AddAnotherPage struct {
	BasePage
	Plan             *domain.Plan // sidebar.html needs .Plan.ID for every nav link
	ItemName         string       // the just-finished item's display name, e.g. `Added "Equipment"`
	Answers          []AnsweredField
	ItemNounSingular string // e.g. "fixed asset" -> "Add another fixed asset?"
	DoneAction       string
	AddAnotherAction string
}

// SectionIntroPage is the generic landing/description page shown when a
// user clicks "Get Started" on a not-yet-started section, used across all
// four hubs via the shared "wizard-section-intro.html" template.
//
// HubBadge/SectionTitle (not Title) deliberately, matching
// QuestionStepPage's naming: a plain "Title" field here would shadow the
// embedded BasePage.Title that base.html's <title> tag actually needs
// (Go's field-resolution rule prefers the shallower field), silently
// dropping the "| Business Planning Tool" suffix from the browser tab.
type SectionIntroPage struct {
	BasePage
	Plan          *domain.Plan // sidebar.html needs .Plan.ID for every nav link
	HubBadge      string
	HubSummaryURL string
	SectionIcon   string
	SectionTitle  string
	Description   string
	ContinueURL   string
}

// RenderSectionListPage renders any hub/entity's repeatable-section list
// page via the single shared template.
func RenderSectionListPage(w http.ResponseWriter, tc map[string]*template.Template, page SectionListPage) {
	renderTemplate(w, tc, "wizard-section-list.html", "base", page)
}

// RenderAddAnotherPage renders any hub/entity's "add another?" interstitial
// via the single shared template.
func RenderAddAnotherPage(w http.ResponseWriter, tc map[string]*template.Template, page AddAnotherPage) {
	renderTemplate(w, tc, "wizard-add-another.html", "base", page)
}

// RenderSectionIntroPage renders any hub's section intro page via the
// single shared template.
func RenderSectionIntroPage(w http.ResponseWriter, tc map[string]*template.Template, page SectionIntroPage) {
	renderTemplate(w, tc, "wizard-section-intro.html", "base", page)
}
