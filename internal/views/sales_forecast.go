// Package views - Sales Forecast hub: a summary page
// (/plan/{id}/sales-forecast) listing Products and Sales Growth Curve
// completion status, plus one wizard per section, built on the same shared
// wizard infrastructure Starting Point introduced.
package views

import (
	"fmt"
	"html/template"
	"net/http"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// --- Section metadata ---

var salesForecastSectionMetas = []sectionMeta{
	{
		Key:       domain.SectionProducts,
		Title:     "Products",
		Icon:      "bi-box-seam",
		ShortDesc: "Your product/service lines, pricing, and unit costs.",
		LongDesc: "Products are the individual product or service lines you sell - " +
			"each with its own starting unit volume, sales price, and direct cost " +
			"per unit. We'll walk through them one at a time so your revenue and " +
			"gross margin projections reflect your actual mix.",
	},
	{
		Key:       domain.SectionSalesGrowthCurve,
		Title:     "Sales Growth Curve",
		Icon:      "bi-graph-up-arrow",
		ShortDesc: "How unit sales are expected to grow over time.",
		LongDesc: "The Sales Growth Curve is the unit-sales growth schedule applied " +
			"to every product line above - quarterly rates for Year 1, then Year 2 " +
			"and Year 3 growth rates. Setting this once keeps every product's " +
			"forecast consistent.",
	},
}

func salesForecastSectionMetaByKey(key string) sectionMeta {
	return sectionMetaByKey(salesForecastSectionMetas, key)
}

// IsSalesForecastComplete reports whether every Sales Forecast sub-section
// is complete, given the section-status map the store returns. Used to
// decide whether the sidebar's Sales Forecast nav icon renders filled.
func IsSalesForecastComplete(sectionStatus map[string]bool) bool {
	return hubComplete(sectionStatus, salesForecastSectionMetas)
}

// --- Answered-so-far summaries ---

var productFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.Product) string
}{
	{"name", "Product Name", func(p domain.Product) string { return p.Name }},
	{"month1-units", "Month 1 Units", func(p domain.Product) string { return fmt.Sprintf("%d", p.Month1Units) }},
	{"price", "Price / Unit", func(p domain.Product) string { return formatMoney(p.PricePerUnit) }},
	{"cost", "Cost / Unit", func(p domain.Product) string { return formatMoney(p.CostPerUnit) }},
}

func productAnsweredFields(product domain.Product, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range productFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(product)})
	}
	return out
}

var salesGrowthCurveFieldDefs = []struct {
	Step  string
	Label string
	Value func(domain.SalesGrowthCurve) string
}{
	{"q1", "Q1 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.Year1QuarterlyPercent(0)) }},
	{"q2", "Q2 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.Year1QuarterlyPercent(1)) }},
	{"q3", "Q3 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.Year1QuarterlyPercent(2)) }},
	{"q4", "Q4 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.Year1QuarterlyPercent(3)) }},
	{"growth-yr2", "Year 2 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.FutureYearPercent(0)) }},
	{"growth-yr3", "Year 3 Growth", func(s domain.SalesGrowthCurve) string { return formatPercent(s.FutureYearPercent(1)) }},
}

func salesGrowthCurveAnsweredFields(curve domain.SalesGrowthCurve, uptoStep string) []AnsweredField {
	var out []AnsweredField
	for _, f := range salesGrowthCurveFieldDefs {
		if f.Step == uptoStep {
			break
		}
		out = append(out, AnsweredField{Label: f.Label, Value: f.Value(curve)})
	}
	return out
}

// --- URL helpers ---

// SalesForecastSummaryURL is the Sales Forecast overview page for a plan.
func SalesForecastSummaryURL(planID uuid.UUID) string {
	return fmt.Sprintf("/plan/%s/sales-forecast", planID)
}

// SalesForecastSectionListURL is the Products section's list/entry page.
func SalesForecastSectionListURL(planID uuid.UUID, section string) string {
	return fmt.Sprintf("/plan/%s/sales-forecast/%s", planID, section)
}

// SalesForecastSectionStepURL is one wizard question for a specific
// Product item.
func SalesForecastSectionStepURL(planID, itemID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/sales-forecast/%s/%s/%s", planID, section, itemID, step)
}

// SalesForecastSectionSingletonStepURL is one wizard question for the
// Sales Growth Curve singleton, which has no item ID.
func SalesForecastSectionSingletonStepURL(planID uuid.UUID, section, step string) string {
	return fmt.Sprintf("/plan/%s/sales-forecast/%s/%s", planID, section, step)
}

// --- Page types ---

// ProductItem pairs a Product's own wizard-item ID with its underlying
// domain value, so the list page can render Edit/Delete links without
// Product itself needing to know about wizard/storage identity.
type ProductItem struct {
	ID      uuid.UUID
	Product domain.Product
}

// ProductStepPage renders a single wizard question for one Product item.
type ProductStepPage struct {
	QuestionStepPage
	Product domain.Product // answers so far, for pre-fill on Back/error
}

// SalesGrowthCurveStepPage renders a single wizard question for the plan's
// Sales Growth Curve singleton.
type SalesGrowthCurveStepPage struct {
	QuestionStepPage
	Curve domain.SalesGrowthCurve
}

// --- Builders ---

// BuildSalesForecastSummaryPage creates the Sales Forecast overview page.
func BuildSalesForecastSummaryPage(r *http.Request, user *domain.User, plan *domain.Plan, sectionStatus map[string]bool, productCount int) HubSummaryPage {
	base := BuildBasePage(r, "Sales Forecast | Business Planning Tool", user)
	base.SalesForecastComplete = IsSalesForecastComplete(sectionStatus)

	counts := map[string]int{domain.SectionProducts: productCount}

	sections := make([]HubSectionStatus, len(salesForecastSectionMetas))
	for i, m := range salesForecastSectionMetas {
		sections[i] = HubSectionStatus{
			Key:           m.Key,
			Title:         m.Title,
			Description:   m.ShortDesc,
			Complete:      sectionStatus[m.Key],
			ItemCount:     counts[m.Key],
			EditURL:       SalesForecastSectionListURL(plan.ID(), m.Key),
			GetStartedURL: fmt.Sprintf("/plan/%s/sales-forecast/intro/%s", plan.ID(), m.Key),
		}
	}

	return HubSummaryPage{
		BasePage:       base,
		Plan:           plan,
		StepBadge:      "Step 4 of 9",
		HubTitle:       "Sales Forecast",
		HubDescription: "Project your revenue by defining your product lines, pricing, and unit growth rates.",
		Sections:       sections,
		BackURL:        PayrollSummaryURL(plan.ID()),
		ContinueURL:    fmt.Sprintf("/plan/%s/operating-expenses", plan.ID()),
	}
}

// BuildSalesForecastSectionIntroPage creates the landing/description page
// shown when a user clicks "Get Started" on a Sales Forecast section.
// section must be one of the domain.SectionProducts/SalesGrowthCurve
// constants.
func BuildSalesForecastSectionIntroPage(r *http.Request, user *domain.User, plan *domain.Plan, section string, salesForecastComplete bool) SectionIntroPage {
	meta := salesForecastSectionMetaByKey(section)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.SalesForecastComplete = salesForecastComplete

	return SectionIntroPage{
		BasePage:      base,
		Plan:          plan,
		HubBadge:      "Sales Forecast",
		HubSummaryURL: SalesForecastSummaryURL(plan.ID()),
		SectionTitle:  meta.Title,
		SectionIcon:   meta.Icon,
		Description:   meta.LongDesc,
		ContinueURL:   SalesForecastSectionListURL(plan.ID(), section),
	}
}

// productListItem formats a Product for the shared section-list page.
func productListItem(planID uuid.UUID, item ProductItem) SectionListItem {
	detail := fmt.Sprintf("%d units/mo · %s price · %s cost", item.Product.Month1Units, formatMoney(item.Product.PricePerUnit), formatMoney(item.Product.CostPerUnit))
	return SectionListItem{
		ID:           item.ID,
		Title:        item.Product.Name,
		Detail:       detail,
		EditURL:      SalesForecastSectionStepURL(planID, item.ID, domain.SectionProducts, "name"),
		DeleteAction: fmt.Sprintf("/plan/%s/sales-forecast/products/%s", planID, item.ID),
	}
}

// BuildProductListPage creates the Products section's list page.
func BuildProductListPage(r *http.Request, user *domain.User, plan *domain.Plan, items []ProductItem, complete bool, draftItemID *uuid.UUID, draftStep string, salesForecastComplete bool) SectionListPage {
	base := BuildBasePage(r, "Products | Business Planning Tool", user)
	base.SalesForecastComplete = salesForecastComplete

	listItems := make([]SectionListItem, len(items))
	for i, item := range items {
		listItems[i] = productListItem(plan.ID(), item)
	}

	var draftStepURL, draftDeleteAction string
	if draftItemID != nil {
		draftStepURL = SalesForecastSectionStepURL(plan.ID(), *draftItemID, domain.SectionProducts, draftStep)
		draftDeleteAction = fmt.Sprintf("/plan/%s/sales-forecast/products/%s", plan.ID(), *draftItemID)
	}

	return SectionListPage{
		BasePage:           base,
		Plan:               plan,
		HubBadge:           "Sales Forecast",
		SectionIcon:        "bi-box-seam",
		SectionTitle:       "Products",
		SectionDescription: "Your product/service lines, pricing, and unit costs.",
		EmptyIcon:          "bi-cart-plus",
		EmptyText:          "No product lines yet.",
		ItemNounSingular:   "product",
		AddButtonLabel:     "Add Product",
		Items:              listItems,
		DraftItemID:        draftItemID,
		DraftStepURL:       draftStepURL,
		DraftDeleteAction:  draftDeleteAction,
		AddAction:          fmt.Sprintf("/plan/%s/sales-forecast/products/new", plan.ID()),
		FinishAction:       fmt.Sprintf("/plan/%s/sales-forecast/products/finish", plan.ID()),
		OverviewURL:        SalesForecastSummaryURL(plan.ID()),
		Complete:           complete,
	}
}

// BuildProductStepPage creates a ProductStepPage.
func BuildProductStepPage(r *http.Request, user *domain.User, plan *domain.Plan, itemID uuid.UUID, product domain.Product, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, salesForecastComplete bool) ProductStepPage {
	meta := salesForecastSectionMetaByKey(domain.SectionProducts)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.SalesForecastComplete = salesForecastComplete
	return ProductStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SalesForecastSectionStepURL(plan.ID(), itemID, domain.SectionProducts, step),
			BackURL:         backURL,
			CancelURL:       SalesForecastSectionListURL(plan.ID(), domain.SectionProducts),
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: productAnsweredFields(product, step),
		},
		Product: product,
	}
}

// BuildProductAddAnotherPage creates the Products "add another?"
// interstitial.
func BuildProductAddAnotherPage(r *http.Request, user *domain.User, plan *domain.Plan, product domain.Product, salesForecastComplete bool) AddAnotherPage {
	base := BuildBasePage(r, "Products | Business Planning Tool", user)
	base.SalesForecastComplete = salesForecastComplete
	return AddAnotherPage{
		BasePage:         base,
		Plan:             plan,
		ItemName:         product.Name,
		Answers:          productAnsweredFields(product, ""),
		ItemNounSingular: "product",
		DoneAction:       fmt.Sprintf("/plan/%s/sales-forecast/products/finish", plan.ID()),
		AddAnotherAction: fmt.Sprintf("/plan/%s/sales-forecast/products/new", plan.ID()),
	}
}

// BuildSalesGrowthCurveStepPage creates a SalesGrowthCurveStepPage.
func BuildSalesGrowthCurveStepPage(r *http.Request, user *domain.User, plan *domain.Plan, curve domain.SalesGrowthCurve, step string, stepNumber, totalSteps int, backURL, buttonLabel, errMsg string, salesForecastComplete bool) SalesGrowthCurveStepPage {
	meta := salesForecastSectionMetaByKey(domain.SectionSalesGrowthCurve)
	base := BuildBasePage(r, meta.Title+" | Business Planning Tool", user)
	base.SalesForecastComplete = salesForecastComplete
	return SalesGrowthCurveStepPage{
		QuestionStepPage: QuestionStepPage{
			BasePage:        base,
			Plan:            plan,
			SectionTitle:    meta.Title,
			SectionIcon:     meta.Icon,
			Step:            step,
			StepNumber:      stepNumber,
			TotalSteps:      totalSteps,
			FormAction:      SalesForecastSectionSingletonStepURL(plan.ID(), domain.SectionSalesGrowthCurve, step),
			BackURL:         backURL,
			CancelURL:       SalesForecastSummaryURL(plan.ID()), // Sales Growth Curve has no list page to cancel back to
			ButtonLabel:     buttonLabel,
			ErrorMessage:    errMsg,
			PreviousAnswers: salesGrowthCurveAnsweredFields(curve, step),
		},
		Curve: curve,
	}
}

// --- Renderers ---

// RenderProductStepPage renders a single Product wizard step.
func RenderProductStepPage(w http.ResponseWriter, tc map[string]*template.Template, page ProductStepPage) {
	renderTemplate(w, tc, "sales-forecast-products-step.html", "base", page)
}

// RenderProductStepPageWithStatus renders a Product wizard step with a
// custom status code (used on validation errors).
func RenderProductStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page ProductStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "sales-forecast-products-step.html", "base", page, statusCode)
}

// RenderSalesGrowthCurveStepPage renders a single Sales Growth Curve
// wizard step.
func RenderSalesGrowthCurveStepPage(w http.ResponseWriter, tc map[string]*template.Template, page SalesGrowthCurveStepPage) {
	renderTemplate(w, tc, "sales-forecast-growth-step.html", "base", page)
}

// RenderSalesGrowthCurveStepPageWithStatus renders a Sales Growth Curve
// wizard step with a custom status code (used on validation errors).
func RenderSalesGrowthCurveStepPageWithStatus(w http.ResponseWriter, tc map[string]*template.Template, page SalesGrowthCurveStepPage, statusCode int) {
	renderTemplateWithStatus(w, tc, "sales-forecast-growth-step.html", "base", page, statusCode)
}
