// Package web - Operating Expenses wizard. Unlike Payroll/Sales
// Forecast/Cash Flow, this hub has exactly one repeatable section (a flat
// list of expense line items), so there's no summary or section-intro
// page - /plan/{id}/operating-expenses is directly the list page, the
// same relationship Fixed Assets/Salary Roles/Products have to their own
// list pages, just without a parent hub summary sitting above it.
package web

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

var operatingExpenseSteps = []string{"name", "amount", "growth"}

// OperatingExpensesController handles the Operating Expenses list wizard.
type OperatingExpensesController struct {
	planSvc       ifaces.PlanService
	opExSvc       ifaces.OperatingExpensesService
	templateCache map[string]*template.Template
}

func NewOperatingExpensesController(
	mux *http.ServeMux,
	planSvc ifaces.PlanService,
	opExSvc ifaces.OperatingExpensesService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *OperatingExpensesController {
	c := &OperatingExpensesController{planSvc: planSvc, opExSvc: opExSvc, templateCache: templateCache}
	viewer := accessMW.RequireAccess(domain.Viewer)
	editor := accessMW.RequireAccess(domain.Editor)

	mux.Handle("GET /plan/{id}/operating-expenses", viewer(http.HandlerFunc(c.GetOperatingExpenseList)))
	mux.Handle("POST /plan/{id}/operating-expenses/new", editor(http.HandlerFunc(c.PostOperatingExpenseNew)))
	mux.Handle("POST /plan/{id}/operating-expenses/finish", editor(http.HandlerFunc(c.PostOperatingExpenseFinish)))
	mux.Handle("GET /plan/{id}/operating-expenses/{itemID}/{step}", viewer(http.HandlerFunc(c.GetOperatingExpenseStep)))
	mux.Handle("POST /plan/{id}/operating-expenses/{itemID}/{step}", editor(http.HandlerFunc(c.PostOperatingExpenseStep)))
	mux.Handle("POST /plan/{id}/operating-expenses/{itemID}", editor(http.HandlerFunc(c.PostOperatingExpenseDelete)))

	return c
}

// operatingExpensesComplete reports whether the Operating Expenses
// section is complete for planID. Used to drive the sidebar's Operating
// Expenses nav icon, which is shown on every page, not just Operating
// Expenses' own.
func (c *OperatingExpensesController) operatingExpensesComplete(planID uuid.UUID) bool {
	status, err := c.opExSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load operating expenses section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsOperatingExpensesComplete(status)
}

// GetOperatingExpenseList GET /plan/{id}/operating-expenses
func (c *OperatingExpensesController) GetOperatingExpenseList(w http.ResponseWriter, r *http.Request) {
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

	completeItems, err := c.opExSvc.ListCompleteOperatingExpenses(planID)
	if err != nil {
		log.Printf("Failed to load operating expenses for plan %s: %v", planID, err)
	}
	items := make([]views.OpExpenseItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.OpExpenseItem{ID: it.ID, Cost: it.Cost}
	}

	sectionStatus, err := c.opExSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load operating expenses section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.opExSvc.FindOperatingExpenseDraft(planID); err != nil {
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
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostOperatingExpenseNew POST /plan/{id}/operating-expenses/new
func (c *OperatingExpensesController) PostOperatingExpenseNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.opExSvc.CreateOperatingExpenseDraft(planID)
	if err != nil {
		log.Printf("Failed to create operating expense draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new operating expense. Please try again.")
		return
	}

	http.Redirect(w, r, views.OperatingExpensesStepURL(planID, itemID, operatingExpenseSteps[0]), http.StatusSeeOther)
}

// GetOperatingExpenseStep GET /plan/{id}/operating-expenses/{itemID}/{step}
func (c *OperatingExpensesController) GetOperatingExpenseStep(w http.ResponseWriter, r *http.Request) {
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
	idx := stepIndex(operatingExpenseSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.opExSvc.GetOperatingExpense(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(operatingExpenseSteps, idx); prev != "" {
		backURL = views.OperatingExpensesStepURL(planID, itemID, prev)
	}

	page := views.BuildOperatingExpenseStepPage(r, user, plan, itemID, item.Cost, step, idx+1, len(operatingExpenseSteps), backURL, "Next", "", c.operatingExpensesComplete(planID))
	views.RenderOperatingExpenseStepPage(w, c.templateCache, page)
}

// PostOperatingExpenseStep POST /plan/{id}/operating-expenses/{itemID}/{step}
func (c *OperatingExpensesController) PostOperatingExpenseStep(w http.ResponseWriter, r *http.Request) {
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
	idx := stepIndex(operatingExpenseSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.opExSvc.GetOperatingExpense(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	cost := item.Cost
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(operatingExpenseSteps, idx); prev != "" {
			backURL = views.OperatingExpensesStepURL(planID, itemID, prev)
		}
		page := views.BuildOperatingExpenseStepPage(r, user, plan, itemID, cost, step, idx+1, len(operatingExpenseSteps), backURL, "Next", errMsg, c.operatingExpensesComplete(planID))
		views.RenderOperatingExpenseStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
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

	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
		newCurrentStep = len(operatingExpenseSteps)
	}

	if err := c.opExSvc.SaveOperatingExpenseStep(itemID, cost, newCurrentStep, newStatus); err != nil {
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

	page := views.BuildOperatingExpenseAddAnotherPage(r, user, plan, cost, c.operatingExpensesComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostOperatingExpenseDelete POST /plan/{id}/operating-expenses/{itemID}
func (c *OperatingExpensesController) PostOperatingExpenseDelete(w http.ResponseWriter, r *http.Request) {
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
	if err := c.opExSvc.DeleteOperatingExpense(itemID); err != nil {
		log.Printf("Failed to delete operating expense %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.OperatingExpensesListURL(planID), http.StatusSeeOther)
}

// PostOperatingExpenseFinish POST /plan/{id}/operating-expenses/finish
func (c *OperatingExpensesController) PostOperatingExpenseFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.opExSvc.MarkWizardSectionComplete(planID, domain.SectionOperatingExpenses); err != nil {
		log.Printf("Failed to mark operating expenses complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, "/plan/"+planID.String()+"/cash-flow", http.StatusSeeOther)
}
