// Package handlers - Sales Forecast wizard: a summary page
// (/plan/{id}/sales-forecast) listing each sub-section's completion
// status, plus one multi-step wizard per sub-section, following the exact
// same pattern starting_point.go established.
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

var (
	productSteps          = []string{"name", "month1-units", "price", "cost"}
	salesGrowthCurveSteps = []string{"q1", "q2", "q3", "q4", "growth-yr2", "growth-yr3"}
)

// salesForecastComplete reports whether every Sales Forecast sub-section
// is complete for planID. Used to drive the sidebar's Sales Forecast nav
// icon, which is shown on every page, not just Sales Forecast's own.
func (app *App) salesForecastComplete(planID uuid.UUID) bool {
	status, err := app.SalesForecastSvc.GetHubStatus(planID)
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
func (app *App) GetSalesForecastSectionIntro() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		section := r.PathValue("section")
		if !isSalesForecastSection(section) {
			app.renderErrorPage(w, r, http.StatusNotFound, "That Sales Forecast section doesn't exist.")
			return
		}

		page := views.BuildSalesForecastSectionIntroPage(r, user, plan, section, app.salesForecastComplete(planID))
		views.RenderSectionIntroPage(w, app.TemplateCache, page)
	}
}

// GetSalesForecastSummary GET /plan/{id}/sales-forecast
func (app *App) GetSalesForecastSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		sectionStatus, err := app.SalesForecastSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
			sectionStatus = map[string]bool{}
		}

		products, err := app.SalesForecastSvc.ListCompleteProducts(planID)
		if err != nil {
			log.Printf("Failed to load products for plan %s: %v", planID, err)
		}

		page := views.BuildSalesForecastSummaryPage(r, user, plan, sectionStatus, len(products))
		views.RenderHubSummaryPage(w, app.TemplateCache, page)
	}
}

// --- Products ---

// GetProductList GET /plan/{id}/sales-forecast/products
func (app *App) GetProductList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.SalesForecastSvc.ListCompleteProducts(planID)
		if err != nil {
			log.Printf("Failed to load products for plan %s: %v", planID, err)
		}
		items := make([]views.ProductItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.ProductItem{ID: it.ID, Product: it.Product}
		}

		sectionStatus, err := app.SalesForecastSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.SalesForecastSvc.FindProductDraft(planID); err != nil {
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
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostProductNew POST /plan/{id}/sales-forecast/products/new
func (app *App) PostProductNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.SalesForecastSvc.CreateProductDraft(planID)
		if err != nil {
			log.Printf("Failed to create product draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new product. Please try again.")
			return
		}

		http.Redirect(w, r, views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, productSteps[0]), http.StatusSeeOther)
	}
}

// GetProductStep GET /plan/{id}/sales-forecast/products/{itemID}/{step}
func (app *App) GetProductStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(productSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.SalesForecastSvc.GetProduct(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(productSteps, idx); prev != "" {
			backURL = views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, prev)
		}

		page := views.BuildProductStepPage(r, user, plan, itemID, item.Product, step, idx+1, len(productSteps), backURL, "Next", "", app.salesForecastComplete(planID))
		views.RenderProductStepPage(w, app.TemplateCache, page)
	}
}

// PostProductStep POST /plan/{id}/sales-forecast/products/{itemID}/{step}
func (app *App) PostProductStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(productSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.SalesForecastSvc.GetProduct(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		product := item.Product
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(productSteps, idx); prev != "" {
				backURL = views.SalesForecastSectionStepURL(planID, itemID, domain.SectionProducts, prev)
			}
			page := views.BuildProductStepPage(r, user, plan, itemID, product, step, idx+1, len(productSteps), backURL, "Next", errMsg, app.salesForecastComplete(planID))
			views.RenderProductStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
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

		if finishNow {
			if err := domain.ValidateProduct(product); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
			newCurrentStep = len(productSteps)
		}

		if err := app.SalesForecastSvc.SaveProductStep(itemID, product, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save product step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
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

		page := views.BuildProductAddAnotherPage(r, user, plan, product, app.salesForecastComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostProductDelete POST /plan/{id}/sales-forecast/products/{itemID}
func (app *App) PostProductDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		if err := app.SalesForecastSvc.DeleteProduct(itemID); err != nil {
			log.Printf("Failed to delete product %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.SalesForecastSectionListURL(planID, domain.SectionProducts), http.StatusSeeOther)
	}
}

// PostProductFinish POST /plan/{id}/sales-forecast/products/finish
func (app *App) PostProductFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.SalesForecastSvc.MarkWizardSectionComplete(planID, domain.SectionProducts); err != nil {
			log.Printf("Failed to mark products complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.SalesForecastSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Sales Growth Curve (singleton) ---

// GetSalesGrowthCurveEntry GET /plan/{id}/sales-forecast/sales-growth-curve
// Redirects into the wizard at the right step: the top, for a plan that
// has never completed the Sales Growth Curve, or wherever the user left off.
func (app *App) GetSalesGrowthCurveEntry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		sectionStatus, err := app.SalesForecastSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
		}

		step := salesGrowthCurveSteps[0]
		if !sectionStatus[domain.SectionSalesGrowthCurve] {
			if row, err := app.SalesForecastSvc.GetSalesGrowthCurveRow(planID); err != nil {
				log.Printf("Failed to load sales growth curve for plan %s: %v", planID, err)
			} else if row.CurrentStep > 0 && row.CurrentStep < len(salesGrowthCurveSteps) {
				step = salesGrowthCurveSteps[row.CurrentStep]
			}
		}

		http.Redirect(w, r, views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, step), http.StatusSeeOther)
	}
}

// GetSalesGrowthCurveStep GET /plan/{id}/sales-forecast/sales-growth-curve/{step}
func (app *App) GetSalesGrowthCurveStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(salesGrowthCurveSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.SalesForecastSvc.GetSalesGrowthCurveRow(planID)
		if err != nil {
			log.Printf("Failed to load sales growth curve for plan %s: %v", planID, err)
			row = &repositories.SalesGrowthCurveRow{}
		}

		backURL := ""
		if prev := prevStepName(salesGrowthCurveSteps, idx); prev != "" {
			backURL = views.SalesForecastSectionSingletonStepURL(planID, domain.SectionSalesGrowthCurve, prev)
		}

		page := views.BuildSalesGrowthCurveStepPage(r, user, plan, row.Curve, step, idx+1, len(salesGrowthCurveSteps), backURL, salesGrowthCurveButtonLabel(idx), "", app.salesForecastComplete(planID))
		views.RenderSalesGrowthCurveStepPage(w, app.TemplateCache, page)
	}
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
func (app *App) PostSalesGrowthCurveStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(salesGrowthCurveSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.SalesForecastSvc.GetSalesGrowthCurveRow(planID)
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
			page := views.BuildSalesGrowthCurveStepPage(r, user, plan, curve, step, idx+1, len(salesGrowthCurveSteps), backURL, salesGrowthCurveButtonLabel(idx), errMsg, app.salesForecastComplete(planID))
			views.RenderSalesGrowthCurveStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
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
		if err := app.SalesForecastSvc.SaveSalesGrowthCurveStep(planID, curve, newCurrentStep); err != nil {
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
		if err := app.SalesForecastSvc.MarkWizardSectionComplete(planID, domain.SectionSalesGrowthCurve); err != nil {
			log.Printf("Failed to mark sales growth curve complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.SalesForecastSummaryURL(planID), http.StatusSeeOther)
	}
}
