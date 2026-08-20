// Package web - Sales Forecast wizard: a summary page
// (/plan/{id}/sales-forecast) listing each sub-section's completion
// status, plus one multi-step wizard per sub-section, following the exact
// same pattern starting_point.go established.
package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"uuid"

	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

var (
	productSteps          = []string{"name", "month1-units", "price", "cost"}
	salesGrowthCurveSteps = []string{"q1", "q2", "q3", "q4", "growth-yr2", "growth-yr3"}
)

// SalesForecastController handles the Sales Forecast hub: a summary page
// plus the Products and Sales Growth Curve sub-wizards.
type SalesForecastController struct {
	planSvc          ifaces.PlanService
	salesForecastSvc ifaces.SalesForecastService
	templateCache    map[string]*template.Template
}

func NewSalesForecastController(
	mux *http.ServeMux,
	planSvc ifaces.PlanService,
	salesForecastSvc ifaces.SalesForecastService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *SalesForecastController {
	c := &SalesForecastController{planSvc: planSvc, salesForecastSvc: salesForecastSvc, templateCache: templateCache}
	viewer := accessMW.RequireAccess(domain.Viewer)
	editor := accessMW.RequireAccess(domain.Editor)

	mux.Handle("GET /plan/{id}/sales-forecast", viewer(http.HandlerFunc(c.GetSalesForecastSummary)))
	mux.Handle("GET /plan/{id}/sales-forecast/intro/{section}", viewer(http.HandlerFunc(c.GetSalesForecastSectionIntro)))

	mux.Handle("GET /plan/{id}/sales-forecast/products", viewer(http.HandlerFunc(c.GetProductList)))
	mux.Handle("POST /plan/{id}/sales-forecast/products/new", editor(http.HandlerFunc(c.PostProductNew)))
	mux.Handle("POST /plan/{id}/sales-forecast/products/finish", editor(http.HandlerFunc(c.PostProductFinish)))
	mux.Handle("GET /plan/{id}/sales-forecast/products/{itemID}/{step}", viewer(http.HandlerFunc(c.GetProductStep)))
	mux.Handle("POST /plan/{id}/sales-forecast/products/{itemID}/{step}", editor(http.HandlerFunc(c.PostProductStep)))
	mux.Handle("POST /plan/{id}/sales-forecast/products/{itemID}", editor(http.HandlerFunc(c.PostProductDelete)))

	mux.Handle("GET /plan/{id}/sales-forecast/sales-growth-curve", viewer(http.HandlerFunc(c.GetSalesGrowthCurveEntry)))
	mux.Handle("GET /plan/{id}/sales-forecast/sales-growth-curve/{step}", viewer(http.HandlerFunc(c.GetSalesGrowthCurveStep)))
	mux.Handle("POST /plan/{id}/sales-forecast/sales-growth-curve/{step}", editor(http.HandlerFunc(c.PostSalesGrowthCurveStep)))

	return c
}

// salesForecastComplete reports whether every Sales Forecast sub-section
// is complete for planID. Used to drive the sidebar's Sales Forecast nav
// icon, which is shown on every page, not just Sales Forecast's own.
func (c *SalesForecastController) salesForecastComplete(planID uuid.UUID) bool {
	status, err := c.salesForecastSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsSalesForecastComplete(status)
}

func isSalesForecastSection(section string) bool {
	switch section {
	case domain.SectionProducts, domain.SectionSalesGrowthCurve:
		return true
	default:
		return false
	}
}

// GetSalesForecastSectionIntro GET /plan/{id}/sales-forecast/{section}/intro
func (c *SalesForecastController) GetSalesForecastSectionIntro(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	section := r.PathValue("section")
	if !isSalesForecastSection(section) {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That Sales Forecast section doesn't exist.")
		return
	}

	page := views.BuildSalesForecastSectionIntroPage(r, user, plan, section, c.salesForecastComplete(planID))
	views.RenderSectionIntroPage(w, c.templateCache, page)
}

// GetSalesForecastSummary GET /plan/{id}/sales-forecast
func (c *SalesForecastController) GetSalesForecastSummary(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	sectionStatus, err := c.salesForecastSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
		sectionStatus = map[string]bool{}
	}

	products, err := c.salesForecastSvc.ListCompleteProducts(planID)
	if err != nil {
		log.Printf("Failed to load products for plan %s: %v", planID, err)
	}

	page := views.BuildSalesForecastSummaryPage(r, user, plan, sectionStatus, len(products))
	views.RenderHubSummaryPage(w, c.templateCache, page)
}

// --- Products ---

// GetProductList GET /plan/{id}/sales-forecast/products
func (c *SalesForecastController) GetProductList(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	completeItems, err := c.salesForecastSvc.ListCompleteProducts(planID)
	if err != nil {
		log.Printf("Failed to load products for plan %s: %v", planID, err)
	}
	items := make([]views.ProductItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.ProductItem{ID: it.ID, Product: it.Product}
	}

	sectionStatus, err := c.salesForecastSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.salesForecastSvc.FindProductDraft(planID); err != nil {
		log.Printf("Failed to load product draft for plan %s: %v", planID, err)
	} else if draft != nil {
		id := draft.ID
		draftItemID = &id
		idx := draft.CurrentStep
		if idx < 0 || idx >= len(productSteps) {
			idx = 0
		}
		draftStep = productSteps[idx]
	}

	page := views.BuildProductListPage(r, user, plan, items, sectionStatus[domain.SectionProducts], draftItemID, draftStep, views.IsSalesForecastComplete(sectionStatus))
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostProductNew POST /plan/{id}/sales-forecast/products/new
func (c *SalesForecastController) PostProductNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.salesForecastSvc.CreateProductDraft(planID)
	if err != nil {
		log.Printf("Failed to create product draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new product. Please try again.")
		return
	}

	http.Redirect(w, r, views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, productSteps[0]), http.StatusSeeOther)
}

// GetProductStep GET /plan/{id}/sales-forecast/products/{itemID}/{step}
func (c *SalesForecastController) GetProductStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(productSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.salesForecastSvc.GetProduct(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(productSteps, idx); prev != "" {
		backURL = views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, prev)
	}

	page := views.BuildProductStepPage(r, user, plan, itemID, item.Product, step, idx+1, len(productSteps), backURL, "Next", "", c.salesForecastComplete(planID))
	views.RenderProductStepPage(w, c.templateCache, page)
}

// PostProductStep POST /plan/{id}/sales-forecast/products/{itemID}/{step}
func (c *SalesForecastController) PostProductStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(productSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.salesForecastSvc.GetProduct(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	product := item.Product
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(productSteps, idx); prev != "" {
			backURL = views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, prev)
		}
		page := views.BuildProductStepPage(r, user, plan, itemID, product, step, idx+1, len(productSteps), backURL, "Next", errMsg, c.salesForecastComplete(planID))
		views.RenderProductStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	switch step {
	case "name":
		product.Name = strings.TrimSpace(r.PostForm.Get("name"))
		if product.Name == "" {
			renderStepError("Please enter a name for this product.")
			return
		}
	case "month1-units":
		units, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("month1_units")))
		if err != nil || units < 0 {
			renderStepError("Please enter a valid unit count.")
			return
		}
		product.Month1Units = units
	case "price":
		price, ok := parseStepMoney(r.PostForm.Get("price"))
		if !ok {
			renderStepError("Please enter a valid price.")
			return
		}
		product.PricePerUnit = price
	case "cost":
		cost, ok := parseStepMoney(r.PostForm.Get("cost"))
		if !ok {
			renderStepError("Please enter a valid cost.")
			return
		}
		product.CostPerUnit = cost
	}

	finishNow := step == "cost"

	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
		newCurrentStep = len(productSteps)
	}

	if err := c.salesForecastSvc.SaveProductStep(itemID, product, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(err.Error())
		} else {
			log.Printf("Failed to save product step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
		return
	}

	if !finishNow {
		http.Redirect(w, r, views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, nextStepName(productSteps, idx)), http.StatusSeeOther)
		return
	}

	if wasComplete {
		http.Redirect(w, r, views.SalesForecastSectionListURL(planID, domain.SectionProducts), http.StatusSeeOther)
		return
	}

	page := views.BuildProductAddAnotherPage(r, user, plan, product, c.salesForecastComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostProductDelete POST /plan/{id}/sales-forecast/products/{itemID}
func (c *SalesForecastController) PostProductDelete(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	if err := c.salesForecastSvc.DeleteProduct(itemID); err != nil {
		log.Printf("Failed to delete product %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.SalesForecastSectionListURL(planID, domain.SectionProducts), http.StatusSeeOther)
}

// PostProductFinish POST /plan/{id}/sales-forecast/products/finish
func (c *SalesForecastController) PostProductFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.salesForecastSvc.MarkWizardSectionComplete(planID, domain.SectionProducts); err != nil {
		log.Printf("Failed to mark products complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.SalesForecastSummaryURL(planID), http.StatusSeeOther)
}

// --- Sales Growth Curve (singleton) ---

// GetSalesGrowthCurveEntry GET /plan/{id}/sales-forecast/sales-growth-curve
// Redirects into the wizard at the right step: the top, for a plan that
// has never completed the Sales Growth Curve, or wherever the user left off.
func (c *SalesForecastController) GetSalesGrowthCurveEntry(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	sectionStatus, err := c.salesForecastSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
	}

	step := salesGrowthCurveSteps[0]
	if !sectionStatus[domain.SectionSalesGrowthCurve] {
		if row, err := c.salesForecastSvc.GetSalesGrowthCurveRow(planID); err != nil {
			log.Printf("Failed to load sales growth curve for plan %s: %v", planID, err)
		} else if row.CurrentStep > 0 && row.CurrentStep < len(salesGrowthCurveSteps) {
			step = salesGrowthCurveSteps[row.CurrentStep]
		}
	}

	http.Redirect(w, r, views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, step), http.StatusSeeOther)
}

// GetSalesGrowthCurveStep GET /plan/{id}/sales-forecast/sales-growth-curve/{step}
func (c *SalesForecastController) GetSalesGrowthCurveStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(salesGrowthCurveSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.salesForecastSvc.GetSalesGrowthCurveRow(planID)
	if err != nil {
		log.Printf("Failed to load sales growth curve for plan %s: %v", planID, err)
		row = &repositories.SalesGrowthCurveRow{}
	}

	backURL := ""
	if prev := prevStepName(salesGrowthCurveSteps, idx); prev != "" {
		backURL = views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, prev)
	}

	page := views.BuildSalesGrowthCurveStepPage(r, user, plan, row.Curve, step, idx+1, len(salesGrowthCurveSteps), backURL, salesGrowthCurveButtonLabel(idx), "", c.salesForecastComplete(planID))
	views.RenderSalesGrowthCurveStepPage(w, c.templateCache, page)
}

// salesGrowthCurveButtonLabel is "Finish" on the last of the Sales Growth
// Curve's 6 fixed steps and "Next" everywhere else.
func salesGrowthCurveButtonLabel(idx int) string {
	if idx == len(salesGrowthCurveSteps)-1 {
		return "Finish"
	}
	return "Next"
}

// PostSalesGrowthCurveStep POST /plan/{id}/sales-forecast/sales-growth-curve/{step}
func (c *SalesForecastController) PostSalesGrowthCurveStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(salesGrowthCurveSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.salesForecastSvc.GetSalesGrowthCurveRow(planID)
	if err != nil {
		log.Printf("Failed to load sales growth curve for plan %s: %v", planID, err)
		row = &repositories.SalesGrowthCurveRow{}
	}
	curve := row.Curve

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(salesGrowthCurveSteps, idx); prev != "" {
			backURL = views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, prev)
		}
		page := views.BuildSalesGrowthCurveStepPage(r, user, plan, curve, step, idx+1, len(salesGrowthCurveSteps), backURL, salesGrowthCurveButtonLabel(idx), errMsg, c.salesForecastComplete(planID))
		views.RenderSalesGrowthCurveStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	rate, ok := parseStepPercent(r.PostForm.Get("rate"))
	if !ok {
		renderStepError("Please enter a valid growth rate.")
		return
	}

	if len(curve.FutureYearRates) < 2 {
		curve.FutureYearRates = []float64{0, 0}
	}

	switch step {
	case "q1":
		curve.Year1QuarterlyRates[0] = rate
	case "q2":
		curve.Year1QuarterlyRates[1] = rate
	case "q3":
		curve.Year1QuarterlyRates[2] = rate
	case "q4":
		curve.Year1QuarterlyRates[3] = rate
	case "growth-yr2":
		curve.FutureYearRates[0] = rate
	case "growth-yr3":
		curve.FutureYearRates[1] = rate
	}

	newCurrentStep := idx + 1
	if err := c.salesForecastSvc.SaveSalesGrowthCurveStep(planID, curve, newCurrentStep); err != nil {
		log.Printf("Failed to save sales growth curve step: %v", err)
		renderStepError("An internal database error occurred. Please try again.")
		return
	}

	if idx < len(salesGrowthCurveSteps)-1 {
		http.Redirect(w, r, views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, nextStepName(salesGrowthCurveSteps, idx)), http.StatusSeeOther)
		return
	}

	// Last step: Sales Growth Curve isn't repeatable, so finishing it
	// marks the section complete directly and returns to the summary.
	if err := c.salesForecastSvc.MarkWizardSectionComplete(planID, domain.SectionSalesGrowthCurve); err != nil {
		log.Printf("Failed to mark sales growth curve complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.SalesForecastSummaryURL(planID), http.StatusSeeOther)
}
