// Package handlers - Cash Flow wizard: a summary page
// (/plan/{id}/cash-flow) listing each sub-section's completion status,
// plus one multi-step wizard per sub-section, following the exact same
// pattern starting_point.go established.
package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/store"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

var (
	inventoryPurchaseSteps = []string{"category", "monthly-amount", "growth-yr2", "growth-yr3"}
	distributionSteps      = []string{"name", "monthly-amount", "growth-yr2", "growth-yr3"}
)

// cashFlowComplete reports whether every Cash Flow sub-section is
// complete for planID. Used to drive the sidebar's Cash Flow nav icon,
// which is shown on every page, not just Cash Flow's own.
func (app *App) cashFlowComplete(planID uuid.UUID) bool {
	status, err := app.Store.GetWizardSectionStatus(planID, domain.HubCashFlow)
	if err != nil {
		log.Printf("Failed to load cash flow section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsCashFlowComplete(status)
}

func isCashFlowSection(section string) bool {
	switch section {
	case domain.SectionInventoryPurchases, domain.SectionDistributions:
		return true
	default:
		return false
	}
}

// GetCashFlowSectionIntro GET /plan/{id}/cash-flow/{section}/intro
func (app *App) GetCashFlowSectionIntro() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		section := r.PathValue("section")
		if !isCashFlowSection(section) {
			app.renderErrorPage(w, r, http.StatusNotFound, "That Cash Flow section doesn't exist.")
			return
		}

		page := views.BuildCashFlowSectionIntroPage(r, user, plan, section, app.cashFlowComplete(planID))
		views.RenderSectionIntroPage(w, app.TemplateCache, page)
	}
}

// GetCashFlowSummary GET /plan/{id}/cash-flow
func (app *App) GetCashFlowSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		sectionStatus, err := app.Store.GetWizardSectionStatus(planID, domain.HubCashFlow)
		if err != nil {
			log.Printf("Failed to load cash flow section status for plan %s: %v", planID, err)
			sectionStatus = map[string]bool{}
		}

		inventory, err := app.Store.ListCompleteInventoryPurchases(planID)
		if err != nil {
			log.Printf("Failed to load inventory purchases for plan %s: %v", planID, err)
		}
		distributions, err := app.Store.ListCompleteDistributions(planID)
		if err != nil {
			log.Printf("Failed to load distributions for plan %s: %v", planID, err)
		}

		page := views.BuildCashFlowSummaryPage(r, user, plan, sectionStatus, len(inventory), len(distributions))
		views.RenderHubSummaryPage(w, app.TemplateCache, page)
	}
}

// --- Inventory Purchases ---

// GetInventoryPurchaseList GET /plan/{id}/cash-flow/inventory-purchases
func (app *App) GetInventoryPurchaseList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.Store.ListCompleteInventoryPurchases(planID)
		if err != nil {
			log.Printf("Failed to load inventory purchases for plan %s: %v", planID, err)
		}
		items := make([]views.InventoryPurchaseItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.InventoryPurchaseItem{ID: it.ID, Purchase: it.Purchase}
		}

		sectionStatus, err := app.Store.GetWizardSectionStatus(planID, domain.HubCashFlow)
		if err != nil {
			log.Printf("Failed to load cash flow section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.Store.GetInventoryPurchaseDraft(planID); err != nil {
			log.Printf("Failed to load inventory purchase draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(inventoryPurchaseSteps) {
				idx = 0
			}
			draftStep = inventoryPurchaseSteps[idx]
		}

		page := views.BuildInventoryPurchaseListPage(r, user, plan, items, sectionStatus[domain.SectionInventoryPurchases], draftItemID, draftStep, views.IsCashFlowComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostInventoryPurchaseNew POST /plan/{id}/cash-flow/inventory-purchases/new
func (app *App) PostInventoryPurchaseNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.Store.CreateInventoryPurchaseDraft(planID)
		if err != nil {
			log.Printf("Failed to create inventory purchase draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new inventory purchase. Please try again.")
			return
		}

		http.Redirect(w, r, views.CashFlowSectionStepURL(planID, itemID, domain.SectionInventoryPurchases, inventoryPurchaseSteps[0]), http.StatusSeeOther)
	}
}

// GetInventoryPurchaseStep GET /plan/{id}/cash-flow/inventory-purchases/{itemID}/{step}
func (app *App) GetInventoryPurchaseStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
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
		idx := stepIndex(inventoryPurchaseSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetInventoryPurchase(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(inventoryPurchaseSteps, idx); prev != "" {
			backURL = views.CashFlowSectionStepURL(planID, itemID, domain.SectionInventoryPurchases, prev)
		}

		page := views.BuildInventoryPurchaseStepPage(r, user, plan, itemID, item.Purchase, step, idx+1, len(inventoryPurchaseSteps), backURL, "Next", "", app.cashFlowComplete(planID))
		views.RenderInventoryPurchaseStepPage(w, app.TemplateCache, page)
	}
}

// PostInventoryPurchaseStep POST /plan/{id}/cash-flow/inventory-purchases/{itemID}/{step}
func (app *App) PostInventoryPurchaseStep() http.HandlerFunc {
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
		plan, err := app.Store.Get(planID)
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
		idx := stepIndex(inventoryPurchaseSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetInventoryPurchase(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		purchase := item.Purchase
		wasComplete := item.Status == store.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(inventoryPurchaseSteps, idx); prev != "" {
				backURL = views.CashFlowSectionStepURL(planID, itemID, domain.SectionInventoryPurchases, prev)
			}
			page := views.BuildInventoryPurchaseStepPage(r, user, plan, itemID, purchase, step, idx+1, len(inventoryPurchaseSteps), backURL, "Next", errMsg, app.cashFlowComplete(planID))
			views.RenderInventoryPurchaseStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "category":
			purchase.Category = strings.TrimSpace(r.PostForm.Get("name"))
			if purchase.Category == "" {
				renderStepError("Please enter a category.")
				return
			}
		case "monthly-amount":
			amt, ok := parseStepMoney(r.PostForm.Get("monthly_amount"))
			if !ok {
				renderStepError("Please enter a valid monthly amount.")
				return
			}
			purchase.MonthlyAmount = amt
		case "growth-yr2":
			rate, ok := parseStepPercent(r.PostForm.Get("growth_yr2"))
			if !ok {
				renderStepError("Please enter a valid growth rate.")
				return
			}
			purchase.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{rate, purchase.GrowthAfterYr1.GrowthRatePercent(1) / 100}}
		case "growth-yr3":
			rate, ok := parseStepPercent(r.PostForm.Get("growth_yr3"))
			if !ok {
				renderStepError("Please enter a valid growth rate.")
				return
			}
			purchase.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{purchase.GrowthAfterYr1.GrowthRatePercent(0) / 100, rate}}
		}

		finishNow := step == "growth-yr3"

		if finishNow {
			if err := domain.ValidateInventoryPurchase(purchase); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := store.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = store.StatusComplete
			newCurrentStep = len(inventoryPurchaseSteps)
		}

		if err := app.Store.SaveInventoryPurchaseStep(itemID, purchase, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save inventory purchase step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.CashFlowSectionStepURL(planID, itemID, domain.SectionInventoryPurchases, nextStepName(inventoryPurchaseSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			http.Redirect(w, r, views.CashFlowSectionListURL(planID, domain.SectionInventoryPurchases), http.StatusSeeOther)
			return
		}

		page := views.BuildInventoryPurchaseAddAnotherPage(r, user, plan, purchase, app.cashFlowComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostInventoryPurchaseDelete POST /plan/{id}/cash-flow/inventory-purchases/{itemID}
func (app *App) PostInventoryPurchaseDelete() http.HandlerFunc {
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
		if err := app.Store.DeleteInventoryPurchase(itemID); err != nil {
			log.Printf("Failed to delete inventory purchase %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.CashFlowSectionListURL(planID, domain.SectionInventoryPurchases), http.StatusSeeOther)
	}
}

// PostInventoryPurchaseFinish POST /plan/{id}/cash-flow/inventory-purchases/finish
func (app *App) PostInventoryPurchaseFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.Store.MarkWizardSectionComplete(planID, domain.HubCashFlow, domain.SectionInventoryPurchases); err != nil {
			log.Printf("Failed to mark inventory purchases complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.CashFlowSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Distributions ---

// GetDistributionList GET /plan/{id}/cash-flow/distributions
func (app *App) GetDistributionList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.Store.ListCompleteDistributions(planID)
		if err != nil {
			log.Printf("Failed to load distributions for plan %s: %v", planID, err)
		}
		items := make([]views.DistributionItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.DistributionItem{ID: it.ID, Distribution: it.Distribution}
		}

		sectionStatus, err := app.Store.GetWizardSectionStatus(planID, domain.HubCashFlow)
		if err != nil {
			log.Printf("Failed to load cash flow section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.Store.GetDistributionDraft(planID); err != nil {
			log.Printf("Failed to load distribution draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(distributionSteps) {
				idx = 0
			}
			draftStep = distributionSteps[idx]
		}

		page := views.BuildDistributionListPage(r, user, plan, items, sectionStatus[domain.SectionDistributions], draftItemID, draftStep, views.IsCashFlowComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostDistributionNew POST /plan/{id}/cash-flow/distributions/new
func (app *App) PostDistributionNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.Store.CreateDistributionDraft(planID)
		if err != nil {
			log.Printf("Failed to create distribution draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new distribution. Please try again.")
			return
		}

		http.Redirect(w, r, views.CashFlowSectionStepURL(planID, itemID, domain.SectionDistributions, distributionSteps[0]), http.StatusSeeOther)
	}
}

// GetDistributionStep GET /plan/{id}/cash-flow/distributions/{itemID}/{step}
func (app *App) GetDistributionStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.Store.Get(planID)
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
		idx := stepIndex(distributionSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetDistribution(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(distributionSteps, idx); prev != "" {
			backURL = views.CashFlowSectionStepURL(planID, itemID, domain.SectionDistributions, prev)
		}

		page := views.BuildDistributionStepPage(r, user, plan, itemID, item.Distribution, step, idx+1, len(distributionSteps), backURL, "Next", "", app.cashFlowComplete(planID))
		views.RenderDistributionStepPage(w, app.TemplateCache, page)
	}
}

// PostDistributionStep POST /plan/{id}/cash-flow/distributions/{itemID}/{step}
func (app *App) PostDistributionStep() http.HandlerFunc {
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
		plan, err := app.Store.Get(planID)
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
		idx := stepIndex(distributionSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetDistribution(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		dist := item.Distribution
		wasComplete := item.Status == store.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(distributionSteps, idx); prev != "" {
				backURL = views.CashFlowSectionStepURL(planID, itemID, domain.SectionDistributions, prev)
			}
			page := views.BuildDistributionStepPage(r, user, plan, itemID, dist, step, idx+1, len(distributionSteps), backURL, "Next", errMsg, app.cashFlowComplete(planID))
			views.RenderDistributionStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "name":
			dist.Name = strings.TrimSpace(r.PostForm.Get("name"))
			if dist.Name == "" {
				renderStepError("Please enter a name for this outflow.")
				return
			}
		case "monthly-amount":
			amt, ok := parseStepMoney(r.PostForm.Get("monthly_amount"))
			if !ok {
				renderStepError("Please enter a valid monthly amount.")
				return
			}
			dist.MonthlyAmount = amt
		case "growth-yr2":
			rate, ok := parseStepPercent(r.PostForm.Get("growth_yr2"))
			if !ok {
				renderStepError("Please enter a valid growth rate.")
				return
			}
			dist.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{rate, dist.GrowthAfterYr1.GrowthRatePercent(1) / 100}}
		case "growth-yr3":
			rate, ok := parseStepPercent(r.PostForm.Get("growth_yr3"))
			if !ok {
				renderStepError("Please enter a valid growth rate.")
				return
			}
			dist.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{dist.GrowthAfterYr1.GrowthRatePercent(0) / 100, rate}}
		}

		finishNow := step == "growth-yr3"

		if finishNow {
			if err := domain.ValidateDistribution(dist); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := store.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = store.StatusComplete
			newCurrentStep = len(distributionSteps)
		}

		if err := app.Store.SaveDistributionStep(itemID, dist, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save distribution step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.CashFlowSectionStepURL(planID, itemID, domain.SectionDistributions, nextStepName(distributionSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			http.Redirect(w, r, views.CashFlowSectionListURL(planID, domain.SectionDistributions), http.StatusSeeOther)
			return
		}

		page := views.BuildDistributionAddAnotherPage(r, user, plan, dist, app.cashFlowComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostDistributionDelete POST /plan/{id}/cash-flow/distributions/{itemID}
func (app *App) PostDistributionDelete() http.HandlerFunc {
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
		if err := app.Store.DeleteDistribution(itemID); err != nil {
			log.Printf("Failed to delete distribution %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.CashFlowSectionListURL(planID, domain.SectionDistributions), http.StatusSeeOther)
	}
}

// PostDistributionFinish POST /plan/{id}/cash-flow/distributions/finish
func (app *App) PostDistributionFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.Store.MarkWizardSectionComplete(planID, domain.HubCashFlow, domain.SectionDistributions); err != nil {
			log.Printf("Failed to mark distributions complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.CashFlowSummaryURL(planID), http.StatusSeeOther)
	}
}
