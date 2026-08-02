// Package handlers contains http functions for mux
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/store"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// App holds the dependencies for all handlers
type App struct {
	Store         store.PlanStore
	TemplateCache map[string]*template.Template
}

// NewApp is the constructor to inject dependencies from main.go
func NewApp(s store.PlanStore, tc map[string]*template.Template) *App {
	return &App{
		Store:         s,
		TemplateCache: tc,
	}
}

// renderErrorPage is a helper to render error pages
func (app *App) renderErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	page := views.BuildErrorPage(r, statusCode, message)
	views.RenderErrorPageWithStatus(w, app.TemplateCache, page, statusCode)
}

// NotFound is user for GET requests (Custom 404 Catch-All for bad URLs)
func (app *App) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.renderErrorPage(w, r, http.StatusNotFound, "The URL you entered does not exist. Please check the address and try again.")
	}
}

func (app *App) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			// Unauthenticated access
			page := views.IndexPage{
				BasePage: views.BasePage{
					Title: "Business Planning Tool",
					Path:  r.URL.Path,
					User:  nil,
				},
				Plans: []*domain.Plan{},
			}
			views.RenderIndexPage(w, app.TemplateCache, page)
			return
		}

		// Fetch plans the user has access to
		plans, err := app.Store.GetUserPlans(user.ID())
		if err != nil {
			plans = []*domain.Plan{}
		}

		pendingInvites := app.pendingInvitesForUser(user)

		page := views.BuildIndexPage(r, user, plans, pendingInvites)
		views.RenderIndexPage(w, app.TemplateCache, page)
	}
}

// pendingInvitesForUser loads the pending invites addressed to a user's
// email and enriches each with the plan name and inviter's name for
// display in the landing page onboarding section.
func (app *App) pendingInvitesForUser(user *domain.User) []views.InviteSummary {
	invites, err := app.Store.GetPendingInvitesForEmail(user.Email())
	if err != nil {
		log.Printf("Failed to load pending invites for %s: %v", user.Email(), err)
		return nil
	}

	summaries := make([]views.InviteSummary, 0, len(invites))
	for _, invite := range invites {
		summary := views.InviteSummary{Invite: invite}

		if plan, err := app.Store.Get(invite.PlanID); err == nil {
			summary.PlanName = plan.Name()
		} else {
			summary.PlanName = "Unknown Plan"
		}

		if inviter, err := app.Store.GetUser(invite.InvitedBy); err == nil {
			summary.InviterName = inviter.FullName()
		} else {
			summary.InviterName = "a collaborator"
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

// PostSetup POST /plan/setup (Saves a brand new plan)
func (app *App) PostSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		companyName := r.PostForm.Get("companyName")
		startMonth, _ := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, _ := strconv.Atoi(r.PostForm.Get("startYear"))

		newID := uuid.New()
		plan, err := domain.NewPlan(newID, companyName, startMonth, startYear, user.ID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = app.Store.Save(plan)
		if err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		// Grant owner access to the creator
		if err := app.Store.GrantAccess(newID, user.ID(), domain.Owner); err != nil {
			log.Printf("Failed to grant owner access: %v", err)
			http.Error(w, "Failed to setup plan access", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plan/"+newID.String()+"/starting-point", http.StatusSeeOther)
	}
}

// PostUpdateSetup POST /plan/{id}/setup (Updates an existing plan)
func (app *App) PostUpdateSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		// Group extraction and type conversion together
		companyName := r.PostForm.Get("companyName")
		startMonth, errMonth := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, errYear := strconv.Atoi(r.PostForm.Get("startYear"))

		// Combine the error checks for related fields
		if errMonth != nil || errYear != nil {
			http.Error(w, "Invalid month or year provided", http.StatusBadRequest)
			return
		}

		// Use the error message directly from your domain logic
		if err := plan.ChangeCoreDetails(companyName, startMonth, startYear); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.Store.Save(plan); err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		// Redirect to the exact same URL, forcing a GET request
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

// GetSetup GET /plan/{id}/setup (Loads an existing plan into the form)
func (app *App) GetSetup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
			return
		}

		invites, isOwner := app.invitesAndOwnerStatus(planID, user)

		page := views.BuildSetupPage(r, user, plan, invites, isOwner, app.startingPointComplete(planID))
		views.RenderSetupPage(w, app.TemplateCache, page)
	}
}

// invitesAndOwnerStatus loads every invite sent for a plan (any status) and
// reports whether the current user is the plan's Owner, so the template can
// decide whether to show the "invite a collaborator" form.
func (app *App) invitesAndOwnerStatus(planID uuid.UUID, user *domain.User) ([]*domain.PlanInvite, bool) {
	invites, err := app.Store.GetInvitesForPlan(planID)
	if err != nil {
		log.Printf("Failed to load invites for plan %s: %v", planID, err)
		invites = nil
	}

	isOwner := false
	if user != nil {
		if access, err := app.Store.GetAccess(planID, user.ID()); err == nil {
			isOwner = access.AccessLevel == domain.Owner
		}
	}

	return invites, isOwner
}

func (app *App) GetPayroll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildPayrollPage(r, GetUserFromContext(r), plan, app.startingPointComplete(planID))
		views.RenderPayrollPage(w, app.TemplateCache, page)
	}
}

// PostPayroll POST /plan/{id}/payroll
func (app *App) PostPayroll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm Error: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		user := GetUserFromContext(r)

		renderError := func(errMsg string, statusCode int) {
			w.WriteHeader(statusCode)
			page := views.BuildPayrollPage(r, user, plan, app.startingPointComplete(id))
			page.ErrorMessage = errMsg
			views.RenderPayrollPage(w, app.TemplateCache, page)
		}

		plan.ClearPayroll()

		// Salary roles
		roles := r.PostForm["pr_role[]"]
		isContractor := r.PostForm["pr_is_contractor[]"]
		headcounts := r.PostForm["pr_headcount[]"]
		monthlyPays := r.PostForm["pr_monthly_pay[]"]
		growthYr2 := r.PostForm["pr_growth_yr2[]"]
		growthYr3 := r.PostForm["pr_growth_yr3[]"]

		for i, role := range roles {
			if i >= len(headcounts) || i >= len(monthlyPays) {
				break
			}
			if strings.TrimSpace(role) == "" {
				continue
			}

			headcount, _ := strconv.Atoi(headcounts[i])
			pay, _ := strconv.ParseFloat(monthlyPays[i], 64)

			var yr2, yr3 float64
			if i < len(growthYr2) {
				yr2, _ = strconv.ParseFloat(growthYr2[i], 64)
			}
			if i < len(growthYr3) {
				yr3, _ = strconv.ParseFloat(growthYr3[i], 64)
			}

			contractor := i < len(isContractor) && isContractor[i] == "true"

			err := plan.AddSalaryRole(domain.SalaryRole{
				Role:         role,
				IsContractor: contractor,
				Headcount:    headcount,
				MonthlyPay:   domain.Money(pay),
				GrowthAfterYr1: domain.AnnualGrowth{
					RatesAfterYear1: []float64{yr2 / 100.0, yr3 / 100.0},
				},
			})
			if err != nil {
				renderError(fmt.Sprintf("Error saving role '%s': %v", role, err), http.StatusBadRequest)
				return
			}
		}

		// Benefits
		benTypes := r.PostForm["ben_type[]"]
		benAmounts := r.PostForm["ben_amount[]"]
		benGrowthYr2 := r.PostForm["ben_growth_yr2[]"]
		benGrowthYr3 := r.PostForm["ben_growth_yr3[]"]

		for i, benType := range benTypes {
			if i >= len(benAmounts) {
				break
			}
			if strings.TrimSpace(benType) == "" {
				continue
			}

			amt, _ := strconv.ParseFloat(benAmounts[i], 64)

			var yr2, yr3 float64
			if i < len(benGrowthYr2) {
				yr2, _ = strconv.ParseFloat(benGrowthYr2[i], 64)
			}
			if i < len(benGrowthYr3) {
				yr3, _ = strconv.ParseFloat(benGrowthYr3[i], 64)
			}

			err := plan.AddBenefit(domain.Benefit{
				Type:          benType,
				MonthlyAmount: domain.Money(amt),
				GrowthAfterYr1: domain.AnnualGrowth{
					RatesAfterYear1: []float64{yr2 / 100.0, yr3 / 100.0},
				},
			})
			if err != nil {
				renderError(fmt.Sprintf("Error saving benefit '%s': %v", benType, err), http.StatusBadRequest)
				return
			}
		}

		// Payroll tax rates
		parseRate := func(key string) float64 {
			val, _ := strconv.ParseFloat(r.PostForm.Get(key), 64)
			return val / 100.0
		}

		plan.SetPayrollTaxRates(domain.PayrollTaxRates{
			SocialSecurityRate: parseRate("tax_ss_rate"),
			MedicareRate:       parseRate("tax_medicare_rate"),
			FUTARate:           parseRate("tax_futa_rate"),
			SUTARate:           parseRate("tax_suta_rate"),
		})

		if err := app.Store.Save(plan); err != nil {
			log.Printf("Store Save Error: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func (app *App) GetSalesForecast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildSalesForecastPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderSalesForecastPage(w, app.TemplateCache, page)
	}
}

// PostSalesForecast POST /plan/{id}/sales-forecast
func (app *App) PostSalesForecast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm Error: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		user := GetUserFromContext(r)

		renderError := func(errMsg string, statusCode int) {
			w.WriteHeader(statusCode)
			page := views.BuildSalesForecastPage(r, user, plan, app.startingPointComplete(plan.ID()))
			page.ErrorMessage = errMsg
			views.RenderSalesForecastPage(w, app.TemplateCache, page)
		}

		plan.ClearSalesForecast()

		parsePercent := func(key string) float64 {
			val, _ := strconv.ParseFloat(r.PostForm.Get(key), 64)
			return val / 100.0
		}

		futureYearStrs := r.PostForm["growth_future_years[]"]
		futureYearRates := make([]float64, 0, len(futureYearStrs))
		for _, s := range futureYearStrs {
			val, _ := strconv.ParseFloat(s, 64)
			futureYearRates = append(futureYearRates, val/100.0)
		}

		plan.SetSalesGrowth(domain.SalesGrowthCurve{
			Year1QuarterlyRates: [4]float64{
				parsePercent("growth_y1_q1"),
				parsePercent("growth_y1_q2"),
				parsePercent("growth_y1_q3"),
				parsePercent("growth_y1_q4"),
			},
			FutureYearRates: futureYearRates,
		})

		names := r.PostForm["prod_name[]"]
		units := r.PostForm["prod_units[]"]
		prices := r.PostForm["prod_price[]"]
		cogs := r.PostForm["prod_cogs[]"]

		for i, name := range names {
			if i >= len(units) || i >= len(prices) || i >= len(cogs) {
				break
			}
			if strings.TrimSpace(name) == "" {
				continue
			}

			unitCount, _ := strconv.Atoi(units[i])
			price, _ := strconv.ParseFloat(prices[i], 64)
			cost, _ := strconv.ParseFloat(cogs[i], 64)

			err := plan.AddProduct(domain.Product{
				Name:         name,
				Month1Units:  unitCount,
				PricePerUnit: domain.Money(price),
				CostPerUnit:  domain.Money(cost),
			})
			if err != nil {
				renderError(fmt.Sprintf("Error saving product '%s': %v", name, err), http.StatusBadRequest)
				return
			}
		}

		if err := app.Store.Save(plan); err != nil {
			log.Printf("Store Save Error: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func (app *App) GetOpExpenses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildOpExpensesPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderOpExpensesPage(w, app.TemplateCache, page)
	}
}

// PostOpExpenses POST /plan/{id}/operating-expenses
func (app *App) PostOpExpenses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm Error: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		user := GetUserFromContext(r)

		renderError := func(errMsg string, statusCode int) {
			w.WriteHeader(statusCode)
			page := views.BuildOpExpensesPage(r, user, plan, app.startingPointComplete(plan.ID()))
			page.ErrorMessage = errMsg
			views.RenderOpExpensesPage(w, app.TemplateCache, page)
		}

		plan.ClearOpEx()

		names := r.PostForm["opex_name[]"]
		amounts := r.PostForm["opex_amount[]"]
		growths := r.PostForm["opex_growth[]"]

		for i, name := range names {
			if i >= len(amounts) {
				break
			}
			if strings.TrimSpace(name) == "" {
				continue
			}

			amt, _ := strconv.ParseFloat(amounts[i], 64)

			var growthRate float64
			if i < len(growths) {
				growthRate, _ = strconv.ParseFloat(growths[i], 64)
			}

			growth := domain.GrowthStrategy{Type: domain.FlatGrowth}
			if growthRate > 0 {
				growth = domain.GrowthStrategy{Type: domain.AnnualStepPercent, AnnualRate: growthRate / 100.0}
			}

			if err := plan.AddOpEx(name, domain.Money(amt), growth); err != nil {
				renderError(fmt.Sprintf("Error saving expense '%s': %v", name, err), http.StatusBadRequest)
				return
			}
		}

		if err := app.Store.Save(plan); err != nil {
			log.Printf("Store Save Error: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func (app *App) GetCashFlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildCashFlowPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderCashFlowPage(w, app.TemplateCache, page)
	}
}

// PostCashFlow POST /plan/{id}/cash-flow
func (app *App) PostCashFlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm Error: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		plan, err := app.Store.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		user := GetUserFromContext(r)

		renderError := func(errMsg string, statusCode int) {
			w.WriteHeader(statusCode)
			page := views.BuildCashFlowPage(r, user, plan, app.startingPointComplete(plan.ID()))
			page.ErrorMessage = errMsg
			views.RenderCashFlowPage(w, app.TemplateCache, page)
		}

		plan.ClearCashFlow()

		parseGrowth := func(yr2Strs, yr3Strs []string, i int) domain.AnnualGrowth {
			var yr2, yr3 float64
			if i < len(yr2Strs) {
				yr2, _ = strconv.ParseFloat(yr2Strs[i], 64)
			}
			if i < len(yr3Strs) {
				yr3, _ = strconv.ParseFloat(yr3Strs[i], 64)
			}
			return domain.AnnualGrowth{RatesAfterYear1: []float64{yr2 / 100.0, yr3 / 100.0}}
		}

		// Additional inventory
		invNames := r.PostForm["inv_name[]"]
		invAmounts := r.PostForm["inv_amount[]"]
		invGrowthYr2 := r.PostForm["inv_growth_yr2[]"]
		invGrowthYr3 := r.PostForm["inv_growth_yr3[]"]

		for i, name := range invNames {
			if i >= len(invAmounts) {
				break
			}
			if strings.TrimSpace(name) == "" {
				continue
			}

			amt, _ := strconv.ParseFloat(invAmounts[i], 64)

			err := plan.AddInventoryPurchase(domain.InventoryPurchase{
				Category:       name,
				MonthlyAmount:  domain.Money(amt),
				GrowthAfterYr1: parseGrowth(invGrowthYr2, invGrowthYr3, i),
			})
			if err != nil {
				renderError(fmt.Sprintf("Error saving inventory '%s': %v", name, err), http.StatusBadRequest)
				return
			}
		}

		// Distributions
		distNames := r.PostForm["dist_name[]"]
		distAmounts := r.PostForm["dist_amount[]"]
		distGrowthYr2 := r.PostForm["dist_growth_yr2[]"]
		distGrowthYr3 := r.PostForm["dist_growth_yr3[]"]

		for i, name := range distNames {
			if i >= len(distAmounts) {
				break
			}
			if strings.TrimSpace(name) == "" {
				continue
			}

			amt, _ := strconv.ParseFloat(distAmounts[i], 64)

			err := plan.AddDistribution(domain.Distribution{
				Name:           name,
				MonthlyAmount:  domain.Money(amt),
				GrowthAfterYr1: parseGrowth(distGrowthYr2, distGrowthYr3, i),
			})
			if err != nil {
				renderError(fmt.Sprintf("Error saving distribution '%s': %v", name, err), http.StatusBadRequest)
				return
			}
		}

		if err := app.Store.Save(plan); err != nil {
			log.Printf("Store Save Error: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func (app *App) GetIncomeStatement() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildIncomeStatementPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderIncomeStatementPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetBalanceSheet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildBalanceSheetPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderBalanceSheetPage(w, app.TemplateCache, page)
	}
}

func (app *App) GetAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.Store.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildAnalyticsPage(r, GetUserFromContext(r), plan, app.startingPointComplete(plan.ID()))
		views.RenderAnalyticsPage(w, app.TemplateCache, page)
	}
}

// PostDeletePlan POST /plan/{id}/delete (requires Owner access)
func (app *App) PostDeletePlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		if err := app.Store.Delete(id); err != nil {
			log.Printf("Store Delete Error: %v", err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to delete plan")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// parsePlanID extracts and parses the plan ID from the request path
func parsePlanID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue("id"))
}
