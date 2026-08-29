// Package web - Payroll wizard: a summary page (/plan/{id}/payroll)
// listing each sub-section's completion status, plus one multi-step
// wizard per sub-section, following the exact same pattern
// starting_point.go established (see its package doc comment).
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
	salaryRoleSteps     = []string{"role", "type", "headcount", "monthly-pay", "growth-yr2", "growth-yr3"}
	benefitSteps        = []string{"type", "monthly-amount", "growth-yr2", "growth-yr3"}
	payrollTaxRateSteps = []string{"social-security", "medicare", "futa", "suta"}
)

// PayrollController handles the Payroll hub: a summary page plus the
// Salary Roles, Benefits, and Payroll Tax Rates sub-wizards.
type PayrollController struct {
	planSvc       ifaces.PlanService
	payrollSvc    ifaces.PayrollService
	templateCache map[string]*template.Template
}

func NewPayrollController(
	mux *http.ServeMux,
	planSvc ifaces.PlanService,
	payrollSvc ifaces.PayrollService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *PayrollController {
	c := &PayrollController{planSvc: planSvc, payrollSvc: payrollSvc, templateCache: templateCache}
	viewer := accessMW.RequireAccess(domain.Viewer)
	editor := accessMW.RequireAccess(domain.Editor)

	mux.Handle("GET /plan/{id}/payroll", viewer(http.HandlerFunc(c.GetPayrollSummary)))
	mux.Handle("GET /plan/{id}/payroll/intro/{section}", viewer(http.HandlerFunc(c.GetPayrollSectionIntro)))

	mux.Handle("GET /plan/{id}/payroll/salary-roles", viewer(http.HandlerFunc(c.GetSalaryRoleList)))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/new", editor(http.HandlerFunc(c.PostSalaryRoleNew)))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/finish", editor(http.HandlerFunc(c.PostSalaryRoleFinish)))
	mux.Handle("GET /plan/{id}/payroll/salary-roles/{itemID}/{step}", viewer(http.HandlerFunc(c.GetSalaryRoleStep)))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/{itemID}/{step}", editor(http.HandlerFunc(c.PostSalaryRoleStep)))
	mux.Handle("POST /plan/{id}/payroll/salary-roles/{itemID}", editor(http.HandlerFunc(c.PostSalaryRoleDelete)))

	mux.Handle("GET /plan/{id}/payroll/benefits", viewer(http.HandlerFunc(c.GetBenefitList)))
	mux.Handle("POST /plan/{id}/payroll/benefits/new", editor(http.HandlerFunc(c.PostBenefitNew)))
	mux.Handle("POST /plan/{id}/payroll/benefits/finish", editor(http.HandlerFunc(c.PostBenefitFinish)))
	mux.Handle("GET /plan/{id}/payroll/benefits/{itemID}/{step}", viewer(http.HandlerFunc(c.GetBenefitStep)))
	mux.Handle("POST /plan/{id}/payroll/benefits/{itemID}/{step}", editor(http.HandlerFunc(c.PostBenefitStep)))
	mux.Handle("POST /plan/{id}/payroll/benefits/{itemID}", editor(http.HandlerFunc(c.PostBenefitDelete)))

	mux.Handle("GET /plan/{id}/payroll/payroll-tax-rates", viewer(http.HandlerFunc(c.GetPayrollTaxRatesEntry)))
	mux.Handle("GET /plan/{id}/payroll/payroll-tax-rates/{step}", viewer(http.HandlerFunc(c.GetPayrollTaxRatesStep)))
	mux.Handle("POST /plan/{id}/payroll/payroll-tax-rates/{step}", editor(http.HandlerFunc(c.PostPayrollTaxRatesStep)))

	return c
}

// payrollComplete reports whether every Payroll sub-section is complete
// for planID. Used to drive the sidebar's Payroll nav icon, which is
// shown on every page, not just Payroll's own.
func (c *PayrollController) payrollComplete(planID uuid.UUID) bool {
	status, err := c.payrollSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsPayrollComplete(status)
}

func isPayrollSection(section string) bool {
	switch section {
	case domain.SectionSalaryRoles, domain.SectionBenefits, domain.SectionPayrollTaxRates:
		return true
	default:
		return false
	}
}

// GetPayrollSectionIntro GET /plan/{id}/payroll/{section}/intro
func (c *PayrollController) GetPayrollSectionIntro(w http.ResponseWriter, r *http.Request) {
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
	if !isPayrollSection(section) {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That Payroll section doesn't exist.")
		return
	}

	page := views.BuildPayrollSectionIntroPage(r, user, plan, section, c.payrollComplete(planID))
	views.RenderSectionIntroPage(w, c.templateCache, page)
}

// GetPayrollSummary GET /plan/{id}/payroll
func (c *PayrollController) GetPayrollSummary(w http.ResponseWriter, r *http.Request) {
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

	sectionStatus, err := c.payrollSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		sectionStatus = map[string]bool{}
	}

	salaryRoles, err := c.payrollSvc.ListCompleteSalaryRoles(planID)
	if err != nil {
		log.Printf("Failed to load salary roles for plan %s: %v", planID, err)
	}
	benefits, err := c.payrollSvc.ListCompleteBenefits(planID)
	if err != nil {
		log.Printf("Failed to load benefits for plan %s: %v", planID, err)
	}

	page := views.BuildPayrollSummaryPage(r, user, plan, sectionStatus, len(salaryRoles), len(benefits))
	views.RenderHubSummaryPage(w, c.templateCache, page)
}

// --- Salary Roles ---

// GetSalaryRoleList GET /plan/{id}/payroll/salary-roles
func (c *PayrollController) GetSalaryRoleList(w http.ResponseWriter, r *http.Request) {
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

	completeItems, err := c.payrollSvc.ListCompleteSalaryRoles(planID)
	if err != nil {
		log.Printf("Failed to load salary roles for plan %s: %v", planID, err)
	}
	items := make([]views.SalaryRoleItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.SalaryRoleItem{ID: it.ID, Role: it.Role}
	}

	sectionStatus, err := c.payrollSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.payrollSvc.FindSalaryRoleDraft(planID); err != nil {
		log.Printf("Failed to load salary role draft for plan %s: %v", planID, err)
	} else if draft != nil {
		id := draft.ID
		draftItemID = &id
		idx := draft.CurrentStep
		if idx < 0 || idx >= len(salaryRoleSteps) {
			idx = 0
		}
		draftStep = salaryRoleSteps[idx]
	}

	page := views.BuildSalaryRoleListPage(r, user, plan, items, sectionStatus[domain.SectionSalaryRoles], draftItemID, draftStep, views.IsPayrollComplete(sectionStatus))
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostSalaryRoleNew POST /plan/{id}/payroll/salary-roles/new
func (c *PayrollController) PostSalaryRoleNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.payrollSvc.CreateSalaryRoleDraft(planID)
	if err != nil {
		log.Printf("Failed to create salary role draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new salary role. Please try again.")
		return
	}

	http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, salaryRoleSteps[0]), http.StatusSeeOther)
}

// GetSalaryRoleStep GET /plan/{id}/payroll/salary-roles/{itemID}/{step}
func (c *PayrollController) GetSalaryRoleStep(w http.ResponseWriter, r *http.Request) {
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
	idx := stepIndex(salaryRoleSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.payrollSvc.GetSalaryRole(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(salaryRoleSteps, idx); prev != "" {
		backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, prev)
	}

	page := views.BuildSalaryRoleStepPage(r, user, plan, itemID, item.Role, step, idx+1, len(salaryRoleSteps), backURL, "Next", "", c.payrollComplete(planID))
	views.RenderSalaryRoleStepPage(w, c.templateCache, page)
}

// PostSalaryRoleStep POST /plan/{id}/payroll/salary-roles/{itemID}/{step}
func (c *PayrollController) PostSalaryRoleStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
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
	idx := stepIndex(salaryRoleSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.payrollSvc.GetSalaryRole(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	role := item.Role
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(salaryRoleSteps, idx); prev != "" {
			backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, prev)
		}
		page := views.BuildSalaryRoleStepPage(r, user, plan, itemID, role, step, idx+1, len(salaryRoleSteps), backURL, "Next", errMsg, c.payrollComplete(planID))
		views.RenderSalaryRoleStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	switch step {
	case "role":
		role.Role = strings.TrimSpace(r.PostForm.Get("name"))
		if role.Role == "" {
			renderStepError("Please enter a name for this role.")
			return
		}
	case "type":
		role.IsContractor = r.PostForm.Get("is_contractor") == "true"
	case "headcount":
		headcount, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("headcount")))
		if err != nil || headcount < 1 {
			renderStepError("Please enter a valid headcount.")
			return
		}
		role.Headcount = headcount
	case "monthly-pay":
		pay, ok := parseStepMoney(r.PostForm.Get("monthly_pay"))
		if !ok {
			renderStepError("Please enter a valid monthly pay amount.")
			return
		}
		role.MonthlyPay = pay
	case "growth-yr2":
		rate, ok := parseStepPercent(r.PostForm.Get("growth_yr2"))
		if !ok {
			renderStepError("Please enter a valid growth rate.")
			return
		}
		role.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{rate, role.GrowthAfterYr1.GrowthRatePercent(1) / 100}}
	case "growth-yr3":
		rate, ok := parseStepPercent(r.PostForm.Get("growth_yr3"))
		if !ok {
			renderStepError("Please enter a valid growth rate.")
			return
		}
		role.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{role.GrowthAfterYr1.GrowthRatePercent(0) / 100, rate}}
	}

	finishNow := step == "growth-yr3"

	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
		newCurrentStep = len(salaryRoleSteps)
	}

	if err := c.payrollSvc.SaveSalaryRoleStep(itemID, role, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(safeErrorMessage("PayrollSvc SaveSalaryRoleStep", err))
		} else {
			log.Printf("Failed to save salary role step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
		return
	}

	if !finishNow {
		http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, nextStepName(salaryRoleSteps, idx)), http.StatusSeeOther)
		return
	}

	if wasComplete {
		http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionSalaryRoles), http.StatusSeeOther)
		return
	}

	page := views.BuildSalaryRoleAddAnotherPage(r, user, plan, role, c.payrollComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostSalaryRoleDelete POST /plan/{id}/payroll/salary-roles/{itemID}
func (c *PayrollController) PostSalaryRoleDelete(w http.ResponseWriter, r *http.Request) {
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
	if err := c.payrollSvc.DeleteSalaryRole(itemID); err != nil {
		log.Printf("Failed to delete salary role %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionSalaryRoles), http.StatusSeeOther)
}

// PostSalaryRoleFinish POST /plan/{id}/payroll/salary-roles/finish
func (c *PayrollController) PostSalaryRoleFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.payrollSvc.MarkWizardSectionComplete(planID, domain.SectionSalaryRoles); err != nil {
		log.Printf("Failed to mark salary roles complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
}

// --- Benefits ---

// GetBenefitList GET /plan/{id}/payroll/benefits
func (c *PayrollController) GetBenefitList(w http.ResponseWriter, r *http.Request) {
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

	completeItems, err := c.payrollSvc.ListCompleteBenefits(planID)
	if err != nil {
		log.Printf("Failed to load benefits for plan %s: %v", planID, err)
	}
	items := make([]views.BenefitItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.BenefitItem{ID: it.ID, Benefit: it.Benefit}
	}

	sectionStatus, err := c.payrollSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.payrollSvc.FindBenefitDraft(planID); err != nil {
		log.Printf("Failed to load benefit draft for plan %s: %v", planID, err)
	} else if draft != nil {
		id := draft.ID
		draftItemID = &id
		idx := draft.CurrentStep
		if idx < 0 || idx >= len(benefitSteps) {
			idx = 0
		}
		draftStep = benefitSteps[idx]
	}

	page := views.BuildBenefitListPage(r, user, plan, items, sectionStatus[domain.SectionBenefits], draftItemID, draftStep, views.IsPayrollComplete(sectionStatus))
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostBenefitNew POST /plan/{id}/payroll/benefits/new
func (c *PayrollController) PostBenefitNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.payrollSvc.CreateBenefitDraft(planID)
	if err != nil {
		log.Printf("Failed to create benefit draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new benefit. Please try again.")
		return
	}

	http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, benefitSteps[0]), http.StatusSeeOther)
}

// GetBenefitStep GET /plan/{id}/payroll/benefits/{itemID}/{step}
func (c *PayrollController) GetBenefitStep(w http.ResponseWriter, r *http.Request) {
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
	idx := stepIndex(benefitSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.payrollSvc.GetBenefit(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(benefitSteps, idx); prev != "" {
		backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, prev)
	}

	page := views.BuildBenefitStepPage(r, user, plan, itemID, item.Benefit, step, idx+1, len(benefitSteps), backURL, "Next", "", c.payrollComplete(planID))
	views.RenderBenefitStepPage(w, c.templateCache, page)
}

// PostBenefitStep POST /plan/{id}/payroll/benefits/{itemID}/{step}
func (c *PayrollController) PostBenefitStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
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
	idx := stepIndex(benefitSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.payrollSvc.GetBenefit(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	benefit := item.Benefit
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(benefitSteps, idx); prev != "" {
			backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, prev)
		}
		page := views.BuildBenefitStepPage(r, user, plan, itemID, benefit, step, idx+1, len(benefitSteps), backURL, "Next", errMsg, c.payrollComplete(planID))
		views.RenderBenefitStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	switch step {
	case "type":
		benefit.Type = strings.TrimSpace(r.PostForm.Get("name"))
		if benefit.Type == "" {
			renderStepError("Please enter a name for this benefit.")
			return
		}
	case "monthly-amount":
		amt, ok := parseStepMoney(r.PostForm.Get("monthly_amount"))
		if !ok {
			renderStepError("Please enter a valid monthly amount.")
			return
		}
		benefit.MonthlyAmount = amt
	case "growth-yr2":
		rate, ok := parseStepPercent(r.PostForm.Get("growth_yr2"))
		if !ok {
			renderStepError("Please enter a valid growth rate.")
			return
		}
		benefit.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{rate, benefit.GrowthAfterYr1.GrowthRatePercent(1) / 100}}
	case "growth-yr3":
		rate, ok := parseStepPercent(r.PostForm.Get("growth_yr3"))
		if !ok {
			renderStepError("Please enter a valid growth rate.")
			return
		}
		benefit.GrowthAfterYr1 = domain.AnnualGrowth{RatesAfterYear1: []float64{benefit.GrowthAfterYr1.GrowthRatePercent(0) / 100, rate}}
	}

	finishNow := step == "growth-yr3"

	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
		newCurrentStep = len(benefitSteps)
	}

	if err := c.payrollSvc.SaveBenefitStep(itemID, benefit, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(safeErrorMessage("PayrollSvc SaveBenefitStep", err))
		} else {
			log.Printf("Failed to save benefit step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
		return
	}

	if !finishNow {
		http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, nextStepName(benefitSteps, idx)), http.StatusSeeOther)
		return
	}

	if wasComplete {
		http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionBenefits), http.StatusSeeOther)
		return
	}

	page := views.BuildBenefitAddAnotherPage(r, user, plan, benefit, c.payrollComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostBenefitDelete POST /plan/{id}/payroll/benefits/{itemID}
func (c *PayrollController) PostBenefitDelete(w http.ResponseWriter, r *http.Request) {
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
	if err := c.payrollSvc.DeleteBenefit(itemID); err != nil {
		log.Printf("Failed to delete benefit %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionBenefits), http.StatusSeeOther)
}

// PostBenefitFinish POST /plan/{id}/payroll/benefits/finish
func (c *PayrollController) PostBenefitFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.payrollSvc.MarkWizardSectionComplete(planID, domain.SectionBenefits); err != nil {
		log.Printf("Failed to mark benefits complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
}

// --- Payroll Tax Rates (singleton) ---

// GetPayrollTaxRatesEntry GET /plan/{id}/payroll/payroll-tax-rates
// Redirects into the wizard at the right step: the top, for a plan that
// has never completed Payroll Tax Rates, or wherever the user left off.
func (c *PayrollController) GetPayrollTaxRatesEntry(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	sectionStatus, err := c.payrollSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
	}

	step := payrollTaxRateSteps[0]
	if !sectionStatus[domain.SectionPayrollTaxRates] {
		if row, err := c.payrollSvc.GetPayrollTaxRatesRow(planID); err != nil {
			log.Printf("Failed to load payroll tax rates for plan %s: %v", planID, err)
		} else if row.CurrentStep > 0 && row.CurrentStep < len(payrollTaxRateSteps) {
			step = payrollTaxRateSteps[row.CurrentStep]
		}
	}

	http.Redirect(w, r, views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, step), http.StatusSeeOther)
}

// GetPayrollTaxRatesStep GET /plan/{id}/payroll/payroll-tax-rates/{step}
func (c *PayrollController) GetPayrollTaxRatesStep(w http.ResponseWriter, r *http.Request) {
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
	idx := stepIndex(payrollTaxRateSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.payrollSvc.GetPayrollTaxRatesRow(planID)
	if err != nil {
		log.Printf("Failed to load payroll tax rates for plan %s: %v", planID, err)
		row = &repositories.PayrollTaxRatesRow{}
	}

	backURL := ""
	if prev := prevStepName(payrollTaxRateSteps, idx); prev != "" {
		backURL = views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, prev)
	}

	page := views.BuildPayrollTaxRatesStepPage(r, user, plan, row.Rates, step, idx+1, len(payrollTaxRateSteps), backURL, payrollTaxRatesButtonLabel(idx), "", c.payrollComplete(planID))
	views.RenderPayrollTaxRatesStepPage(w, c.templateCache, page)
}

// payrollTaxRatesButtonLabel is "Finish" on the last of Payroll Tax
// Rates' 4 fixed steps (it isn't repeatable, so there's no "Add another?"
// detour after it) and "Next" everywhere else.
func payrollTaxRatesButtonLabel(idx int) string {
	if idx == len(payrollTaxRateSteps)-1 {
		return "Finish"
	}
	return "Next"
}

// PostPayrollTaxRatesStep POST /plan/{id}/payroll/payroll-tax-rates/{step}
func (c *PayrollController) PostPayrollTaxRatesStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
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
	idx := stepIndex(payrollTaxRateSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.payrollSvc.GetPayrollTaxRatesRow(planID)
	if err != nil {
		log.Printf("Failed to load payroll tax rates for plan %s: %v", planID, err)
		row = &repositories.PayrollTaxRatesRow{}
	}
	rates := row.Rates

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(payrollTaxRateSteps, idx); prev != "" {
			backURL = views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, prev)
		}
		page := views.BuildPayrollTaxRatesStepPage(r, user, plan, rates, step, idx+1, len(payrollTaxRateSteps), backURL, payrollTaxRatesButtonLabel(idx), errMsg, c.payrollComplete(planID))
		views.RenderPayrollTaxRatesStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	rate, ok := parseStepPercent(r.PostForm.Get("rate"))
	if !ok || rate < 0 {
		renderStepError("Please enter a valid rate.")
		return
	}

	switch step {
	case "social-security":
		rates.SocialSecurityRate = rate
	case "medicare":
		rates.MedicareRate = rate
	case "futa":
		rates.FUTARate = rate
	case "suta":
		rates.SUTARate = rate
	}

	newCurrentStep := idx + 1
	isLastStep := idx == len(payrollTaxRateSteps)-1
	if err := c.payrollSvc.SavePayrollTaxRatesStep(planID, rates, newCurrentStep, isLastStep); err != nil {
		if isLastStep {
			renderStepError(safeErrorMessage("PayrollSvc SavePayrollTaxRatesStep", err))
		} else {
			log.Printf("Failed to save payroll tax rates step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
		return
	}

	if idx < len(payrollTaxRateSteps)-1 {
		http.Redirect(w, r, views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, nextStepName(payrollTaxRateSteps, idx)), http.StatusSeeOther)
		return
	}

	// Last step: Payroll Tax Rates isn't repeatable, so finishing it
	// marks the section complete directly and returns to the summary.
	if err := c.payrollSvc.MarkWizardSectionComplete(planID, domain.SectionPayrollTaxRates); err != nil {
		log.Printf("Failed to mark payroll tax rates complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
}
