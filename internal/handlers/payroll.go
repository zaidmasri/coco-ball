// Package handlers - Payroll wizard: a summary page (/plan/{id}/payroll)
// listing each sub-section's completion status, plus one multi-step
// wizard per sub-section, following the exact same pattern
// starting_point.go established (see its package doc comment).
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

var (
	salaryRoleSteps     = []string{"role", "type", "headcount", "monthly-pay", "growth-yr2", "growth-yr3"}
	benefitSteps        = []string{"type", "monthly-amount", "growth-yr2", "growth-yr3"}
	payrollTaxRateSteps = []string{"social-security", "medicare", "futa", "suta"}
)

// payrollComplete reports whether every Payroll sub-section is complete
// for planID. Used to drive the sidebar's Payroll nav icon, which is
// shown on every page, not just Payroll's own.
func (app *App) payrollComplete(planID uuid.UUID) bool {
	status, err := app.PayrollSvc.GetHubStatus(planID)
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
func (app *App) GetPayrollSectionIntro() http.HandlerFunc {
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
		if !isPayrollSection(section) {
			app.renderErrorPage(w, r, http.StatusNotFound, "That Payroll section doesn't exist.")
			return
		}

		page := views.BuildPayrollSectionIntroPage(r, user, plan, section, app.payrollComplete(planID))
		views.RenderSectionIntroPage(w, app.TemplateCache, page)
	}
}

// GetPayrollSummary GET /plan/{id}/payroll
func (app *App) GetPayrollSummary() http.HandlerFunc {
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

		sectionStatus, err := app.PayrollSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
			sectionStatus = map[string]bool{}
		}

		salaryRoles, err := app.PayrollSvc.ListCompleteSalaryRoles(planID)
		if err != nil {
			log.Printf("Failed to load salary roles for plan %s: %v", planID, err)
		}
		benefits, err := app.PayrollSvc.ListCompleteBenefits(planID)
		if err != nil {
			log.Printf("Failed to load benefits for plan %s: %v", planID, err)
		}

		page := views.BuildPayrollSummaryPage(r, user, plan, sectionStatus, len(salaryRoles), len(benefits))
		views.RenderHubSummaryPage(w, app.TemplateCache, page)
	}
}

// --- Salary Roles ---

// GetSalaryRoleList GET /plan/{id}/payroll/salary-roles
func (app *App) GetSalaryRoleList() http.HandlerFunc {
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

		completeItems, err := app.PayrollSvc.ListCompleteSalaryRoles(planID)
		if err != nil {
			log.Printf("Failed to load salary roles for plan %s: %v", planID, err)
		}
		items := make([]views.SalaryRoleItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.SalaryRoleItem{ID: it.ID, Role: it.Role}
		}

		sectionStatus, err := app.PayrollSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.PayrollSvc.GetSalaryRoleDraft(planID); err != nil {
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
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostSalaryRoleNew POST /plan/{id}/payroll/salary-roles/new
func (app *App) PostSalaryRoleNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.PayrollSvc.CreateSalaryRoleDraft(planID)
		if err != nil {
			log.Printf("Failed to create salary role draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new salary role. Please try again.")
			return
		}

		http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, salaryRoleSteps[0]), http.StatusSeeOther)
	}
}

// GetSalaryRoleStep GET /plan/{id}/payroll/salary-roles/{itemID}/{step}
func (app *App) GetSalaryRoleStep() http.HandlerFunc {
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
		idx := stepIndex(salaryRoleSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.PayrollSvc.GetSalaryRole(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(salaryRoleSteps, idx); prev != "" {
			backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, prev)
		}

		page := views.BuildSalaryRoleStepPage(r, user, plan, itemID, item.Role, step, idx+1, len(salaryRoleSteps), backURL, "Next", "", app.payrollComplete(planID))
		views.RenderSalaryRoleStepPage(w, app.TemplateCache, page)
	}
}

// PostSalaryRoleStep POST /plan/{id}/payroll/salary-roles/{itemID}/{step}
func (app *App) PostSalaryRoleStep() http.HandlerFunc {
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
		idx := stepIndex(salaryRoleSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.PayrollSvc.GetSalaryRole(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		role := item.Role
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(salaryRoleSteps, idx); prev != "" {
				backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionSalaryRoles, prev)
			}
			page := views.BuildSalaryRoleStepPage(r, user, plan, itemID, role, step, idx+1, len(salaryRoleSteps), backURL, "Next", errMsg, app.payrollComplete(planID))
			views.RenderSalaryRoleStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
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

		if finishNow {
			if err := domain.ValidateSalaryRole(role); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
			newCurrentStep = len(salaryRoleSteps)
		}

		if err := app.PayrollSvc.SaveSalaryRoleStep(itemID, role, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save salary role step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
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

		page := views.BuildSalaryRoleAddAnotherPage(r, user, plan, role, app.payrollComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostSalaryRoleDelete POST /plan/{id}/payroll/salary-roles/{itemID}
func (app *App) PostSalaryRoleDelete() http.HandlerFunc {
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
		if err := app.PayrollSvc.DeleteSalaryRole(itemID); err != nil {
			log.Printf("Failed to delete salary role %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionSalaryRoles), http.StatusSeeOther)
	}
}

// PostSalaryRoleFinish POST /plan/{id}/payroll/salary-roles/finish
func (app *App) PostSalaryRoleFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.PayrollSvc.MarkWizardSectionComplete(planID, domain.SectionSalaryRoles); err != nil {
			log.Printf("Failed to mark salary roles complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Benefits ---

// GetBenefitList GET /plan/{id}/payroll/benefits
func (app *App) GetBenefitList() http.HandlerFunc {
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

		completeItems, err := app.PayrollSvc.ListCompleteBenefits(planID)
		if err != nil {
			log.Printf("Failed to load benefits for plan %s: %v", planID, err)
		}
		items := make([]views.BenefitItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.BenefitItem{ID: it.ID, Benefit: it.Benefit}
		}

		sectionStatus, err := app.PayrollSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.PayrollSvc.GetBenefitDraft(planID); err != nil {
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
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostBenefitNew POST /plan/{id}/payroll/benefits/new
func (app *App) PostBenefitNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.PayrollSvc.CreateBenefitDraft(planID)
		if err != nil {
			log.Printf("Failed to create benefit draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new benefit. Please try again.")
			return
		}

		http.Redirect(w, r, views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, benefitSteps[0]), http.StatusSeeOther)
	}
}

// GetBenefitStep GET /plan/{id}/payroll/benefits/{itemID}/{step}
func (app *App) GetBenefitStep() http.HandlerFunc {
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
		idx := stepIndex(benefitSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.PayrollSvc.GetBenefit(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(benefitSteps, idx); prev != "" {
			backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, prev)
		}

		page := views.BuildBenefitStepPage(r, user, plan, itemID, item.Benefit, step, idx+1, len(benefitSteps), backURL, "Next", "", app.payrollComplete(planID))
		views.RenderBenefitStepPage(w, app.TemplateCache, page)
	}
}

// PostBenefitStep POST /plan/{id}/payroll/benefits/{itemID}/{step}
func (app *App) PostBenefitStep() http.HandlerFunc {
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
		idx := stepIndex(benefitSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.PayrollSvc.GetBenefit(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		benefit := item.Benefit
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(benefitSteps, idx); prev != "" {
				backURL = views.PayrollSectionStepURL(planID, itemID, domain.SectionBenefits, prev)
			}
			page := views.BuildBenefitStepPage(r, user, plan, itemID, benefit, step, idx+1, len(benefitSteps), backURL, "Next", errMsg, app.payrollComplete(planID))
			views.RenderBenefitStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
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

		if finishNow {
			if err := domain.ValidateBenefit(benefit); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
			newCurrentStep = len(benefitSteps)
		}

		if err := app.PayrollSvc.SaveBenefitStep(itemID, benefit, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save benefit step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
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

		page := views.BuildBenefitAddAnotherPage(r, user, plan, benefit, app.payrollComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostBenefitDelete POST /plan/{id}/payroll/benefits/{itemID}
func (app *App) PostBenefitDelete() http.HandlerFunc {
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
		if err := app.PayrollSvc.DeleteBenefit(itemID); err != nil {
			log.Printf("Failed to delete benefit %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.PayrollSectionListURL(planID, domain.SectionBenefits), http.StatusSeeOther)
	}
}

// PostBenefitFinish POST /plan/{id}/payroll/benefits/finish
func (app *App) PostBenefitFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.PayrollSvc.MarkWizardSectionComplete(planID, domain.SectionBenefits); err != nil {
			log.Printf("Failed to mark benefits complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Payroll Tax Rates (singleton) ---

// GetPayrollTaxRatesEntry GET /plan/{id}/payroll/payroll-tax-rates
// Redirects into the wizard at the right step: the top, for a plan that
// has never completed Payroll Tax Rates, or wherever the user left off.
func (app *App) GetPayrollTaxRatesEntry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		sectionStatus, err := app.PayrollSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		}

		step := payrollTaxRateSteps[0]
		if !sectionStatus[domain.SectionPayrollTaxRates] {
			if row, err := app.PayrollSvc.GetPayrollTaxRatesRow(planID); err != nil {
				log.Printf("Failed to load payroll tax rates for plan %s: %v", planID, err)
			} else if row.CurrentStep > 0 && row.CurrentStep < len(payrollTaxRateSteps) {
				step = payrollTaxRateSteps[row.CurrentStep]
			}
		}

		http.Redirect(w, r, views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, step), http.StatusSeeOther)
	}
}

// GetPayrollTaxRatesStep GET /plan/{id}/payroll/payroll-tax-rates/{step}
func (app *App) GetPayrollTaxRatesStep() http.HandlerFunc {
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
		idx := stepIndex(payrollTaxRateSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.PayrollSvc.GetPayrollTaxRatesRow(planID)
		if err != nil {
			log.Printf("Failed to load payroll tax rates for plan %s: %v", planID, err)
			row = &repositories.PayrollTaxRatesRow{}
		}

		backURL := ""
		if prev := prevStepName(payrollTaxRateSteps, idx); prev != "" {
			backURL = views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, prev)
		}

		page := views.BuildPayrollTaxRatesStepPage(r, user, plan, row.Rates, step, idx+1, len(payrollTaxRateSteps), backURL, payrollTaxRatesButtonLabel(idx), "", app.payrollComplete(planID))
		views.RenderPayrollTaxRatesStepPage(w, app.TemplateCache, page)
	}
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
func (app *App) PostPayrollTaxRatesStep() http.HandlerFunc {
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
		idx := stepIndex(payrollTaxRateSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.PayrollSvc.GetPayrollTaxRatesRow(planID)
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
			page := views.BuildPayrollTaxRatesStepPage(r, user, plan, rates, step, idx+1, len(payrollTaxRateSteps), backURL, payrollTaxRatesButtonLabel(idx), errMsg, app.payrollComplete(planID))
			views.RenderPayrollTaxRatesStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
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
		if err := app.PayrollSvc.SavePayrollTaxRatesStep(planID, rates, newCurrentStep); err != nil {
			log.Printf("Failed to save payroll tax rates step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if idx < len(payrollTaxRateSteps)-1 {
			http.Redirect(w, r, views.PayrollSectionSingletonStepURL(planID, domain.SectionPayrollTaxRates, nextStepName(payrollTaxRateSteps, idx)), http.StatusSeeOther)
			return
		}

		// Last step: Payroll Tax Rates isn't repeatable, so finishing it
		// marks the section complete directly and returns to the summary.
		if err := app.PayrollSvc.MarkWizardSectionComplete(planID, domain.SectionPayrollTaxRates); err != nil {
			log.Printf("Failed to mark payroll tax rates complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.PayrollSummaryURL(planID), http.StatusSeeOther)
	}
}
