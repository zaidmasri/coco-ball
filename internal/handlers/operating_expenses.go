// Package handlers - Operating Expenses wizard. Unlike Payroll/Sales
// Forecast/Cash Flow, this hub has exactly one repeatable section (a flat
// list of expense line items), so there's no summary or section-intro
// page - /plan/{id}/operating-expenses is directly the list page, the
// same relationship Fixed Assets/Salary Roles/Products have to their own
// list pages, just without a parent hub summary sitting above it.
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

var operatingExpenseSteps = []string{"name", "amount", "growth"}

// operatingExpensesComplete reports whether the Operating Expenses
// section is complete for planID. Used to drive the sidebar's Operating
// Expenses nav icon, which is shown on every page, not just Operating
// Expenses' own.
func (app *App) operatingExpensesComplete(planID uuid.UUID) bool {
	status, err := app.Store.GetWizardSectionStatus(planID, domain.HubOperatingExpenses)
	if err != nil {
		log.Printf("Failed to load operating expenses section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsOperatingExpensesComplete(status)
}

// GetOperatingExpenseList GET /plan/{id}/operating-expenses
func (app *App) GetOperatingExpenseList() http.HandlerFunc {
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

		completeItems, err := app.Store.ListCompleteOperatingExpenses(planID)
		if err != nil {
			log.Printf("Failed to load operating expenses for plan %s: %v", planID, err)
		}
		items := make([]views.OpExpenseItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.OpExpenseItem{ID: it.ID, Cost: it.Cost}
		}

		sectionStatus, err := app.Store.GetWizardSectionStatus(planID, domain.HubOperatingExpenses)
		if err != nil {
			log.Printf("Failed to load operating expenses section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.Store.GetOperatingExpenseDraft(planID); err != nil {
			log.Printf("Failed to load operating expense draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(operatingExpenseSteps) {
				idx = 0
			}
			draftStep = operatingExpenseSteps[idx]
		}

		page := views.BuildOperatingExpenseListPage(r, user, plan, items, sectionStatus[domain.SectionOperatingExpenses], draftItemID, draftStep, views.IsOperatingExpensesComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostOperatingExpenseNew POST /plan/{id}/operating-expenses/new
func (app *App) PostOperatingExpenseNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.Store.CreateOperatingExpenseDraft(planID)
		if err != nil {
			log.Printf("Failed to create operating expense draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new operating expense. Please try again.")
			return
		}

		http.Redirect(w, r, views.OperatingExpensesStepURL(planID, itemID, operatingExpenseSteps[0]), http.StatusSeeOther)
	}
}

// GetOperatingExpenseStep GET /plan/{id}/operating-expenses/{itemID}/{step}
func (app *App) GetOperatingExpenseStep() http.HandlerFunc {
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
		idx := stepIndex(operatingExpenseSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetOperatingExpense(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(operatingExpenseSteps, idx); prev != "" {
			backURL = views.OperatingExpensesStepURL(planID, itemID, prev)
		}

		page := views.BuildOperatingExpenseStepPage(r, user, plan, itemID, item.Cost, step, idx+1, len(operatingExpenseSteps), backURL, "Next", "", app.operatingExpensesComplete(planID))
		views.RenderOperatingExpenseStepPage(w, app.TemplateCache, page)
	}
}

// PostOperatingExpenseStep POST /plan/{id}/operating-expenses/{itemID}/{step}
func (app *App) PostOperatingExpenseStep() http.HandlerFunc {
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
		idx := stepIndex(operatingExpenseSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.Store.GetOperatingExpense(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		cost := item.Cost
		wasComplete := item.Status == store.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(operatingExpenseSteps, idx); prev != "" {
				backURL = views.OperatingExpensesStepURL(planID, itemID, prev)
			}
			page := views.BuildOperatingExpenseStepPage(r, user, plan, itemID, cost, step, idx+1, len(operatingExpenseSteps), backURL, "Next", errMsg, app.operatingExpensesComplete(planID))
			views.RenderOperatingExpenseStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "name":
			cost.Name = strings.TrimSpace(r.PostForm.Get("name"))
			if cost.Name == "" {
				renderStepError("Please enter a name for this expense.")
				return
			}
		case "amount":
			amt, ok := parseStepMoney(r.PostForm.Get("amount"))
			if !ok {
				renderStepError("Please enter a valid monthly amount.")
				return
			}
			cost.BaseAmountPerMonth = amt
		case "growth":
			rate, ok := parseStepPercent(r.PostForm.Get("growth"))
			if !ok {
				renderStepError("Please enter a valid growth rate.")
				return
			}
			// A 0% rate is stored as flat (no compounding math needed);
			// anything else steps up annually - matches the previous
			// single-page form's behavior of deriving the strategy from
			// the rate itself rather than asking a separate question.
			if rate > 0 {
				cost.Growth = domain.GrowthStrategy{Type: domain.AnnualStepPercent, AnnualRate: rate}
			} else {
				cost.Growth = domain.GrowthStrategy{Type: domain.FlatGrowth}
			}
		}

		finishNow := step == "growth"

		if finishNow {
			if err := domain.ValidateOpExpense(cost); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := store.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = store.StatusComplete
			newCurrentStep = len(operatingExpenseSteps)
		}

		if err := app.Store.SaveOperatingExpenseStep(itemID, cost, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save operating expense step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.OperatingExpensesStepURL(planID, itemID, nextStepName(operatingExpenseSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			http.Redirect(w, r, views.OperatingExpensesListURL(planID), http.StatusSeeOther)
			return
		}

		page := views.BuildOperatingExpenseAddAnotherPage(r, user, plan, cost, app.operatingExpensesComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostOperatingExpenseDelete POST /plan/{id}/operating-expenses/{itemID}
func (app *App) PostOperatingExpenseDelete() http.HandlerFunc {
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
		if err := app.Store.DeleteOperatingExpense(itemID); err != nil {
			log.Printf("Failed to delete operating expense %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.OperatingExpensesListURL(planID), http.StatusSeeOther)
	}
}

// PostOperatingExpenseFinish POST /plan/{id}/operating-expenses/finish
func (app *App) PostOperatingExpenseFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.Store.MarkWizardSectionComplete(planID, domain.HubOperatingExpenses, domain.SectionOperatingExpenses); err != nil {
			log.Printf("Failed to mark operating expenses complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, "/plan/"+planID.String()+"/cash-flow", http.StatusSeeOther)
	}
}
