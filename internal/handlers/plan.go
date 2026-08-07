// Package handlers contains http functions for mux
package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// App holds the dependencies for all handlers
type App struct {
	PlanSvc          ifaces.PlanService
	AuthSvc          ifaces.AuthService
	AccessSvc        ifaces.AccessService
	InviteSvc        ifaces.InviteService
	StartingPointSvc ifaces.StartingPointService
	PayrollSvc       ifaces.PayrollService
	SalesForecastSvc ifaces.SalesForecastService
	CashFlowSvc      ifaces.CashFlowService
	OpExSvc          ifaces.OperatingExpensesService
	HubSvc           ifaces.HubCompletionService
	TemplateCache    map[string]*template.Template
}

// NewApp constructs the App with all service dependencies.
func NewApp(
	planSvc ifaces.PlanService,
	authSvc ifaces.AuthService,
	accessSvc ifaces.AccessService,
	inviteSvc ifaces.InviteService,
	startingPointSvc ifaces.StartingPointService,
	payrollSvc ifaces.PayrollService,
	salesForecastSvc ifaces.SalesForecastService,
	cashFlowSvc ifaces.CashFlowService,
	opExSvc ifaces.OperatingExpensesService,
	hubSvc ifaces.HubCompletionService,
	tc map[string]*template.Template,
) *App {
	return &App{
		PlanSvc:          planSvc,
		AuthSvc:          authSvc,
		AccessSvc:        accessSvc,
		InviteSvc:        inviteSvc,
		StartingPointSvc: startingPointSvc,
		PayrollSvc:       payrollSvc,
		SalesForecastSvc: salesForecastSvc,
		CashFlowSvc:      cashFlowSvc,
		OpExSvc:          opExSvc,
		HubSvc:           hubSvc,
		TemplateCache:    tc,
	}
}

// renderErrorPage is a helper to render error pages
func (app *App) renderErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	page := views.BuildErrorPage(r, statusCode, message)
	views.RenderErrorPageWithStatus(w, app.TemplateCache, page, statusCode)
}

// toViewsHubCompletion translates the application-layer HubCompletion to the
// views-layer equivalent (same fields, different package).
func toViewsHubCompletion(c interfaces.HubCompletion) views.HubCompletion {
	return views.HubCompletion{
		StartingPoint:     c.StartingPoint,
		Payroll:           c.Payroll,
		SalesForecast:     c.SalesForecast,
		OperatingExpenses: c.OperatingExpenses,
		CashFlow:          c.CashFlow,
	}
}

// NotFound is used for GET requests (Custom 404 Catch-All for bad URLs)
func (app *App) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.renderErrorPage(w, r, http.StatusNotFound, "The URL you entered does not exist. Please check the address and try again.")
	}
}

func (app *App) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
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

		plans, err := app.PlanSvc.GetUserPlans(user.ID())
		if err != nil {
			plans = []*domain.Plan{}
		}

		pendingInvites := app.pendingInvitesForUser(user)

		page := views.BuildIndexPage(r, user, plans, pendingInvites)
		views.RenderIndexPage(w, app.TemplateCache, page)
	}
}

// pendingInvitesForUser loads pending invites for the user and enriches each
// with the plan name and inviter name for the landing page onboarding section.
func (app *App) pendingInvitesForUser(user *domain.User) []views.InviteSummary {
	invites, err := app.InviteSvc.GetPendingInvitesForEmail(user.Email())
	if err != nil {
		log.Printf("Failed to load pending invites for %s: %v", user.Email(), err)
		return nil
	}

	summaries := make([]views.InviteSummary, 0, len(invites))
	for _, invite := range invites {
		summary := views.InviteSummary{Invite: invite}

		if plan, err := app.PlanSvc.Get(invite.PlanID); err == nil {
			summary.PlanName = plan.Name()
		} else {
			summary.PlanName = "Unknown Plan"
		}

		if inviter, err := app.AuthSvc.GetUser(invite.InvitedBy); err == nil {
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

		if err = app.PlanSvc.Save(plan); err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

		if err := app.AccessSvc.GrantAccess(newID, user.ID(), domain.Owner); err != nil {
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

		plan, err := app.PlanSvc.Get(id)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		companyName := r.PostForm.Get("companyName")
		startMonth, errMonth := strconv.Atoi(r.PostForm.Get("startMonth"))
		startYear, errYear := strconv.Atoi(r.PostForm.Get("startYear"))

		if errMonth != nil || errYear != nil {
			http.Error(w, "Invalid month or year provided", http.StatusBadRequest)
			return
		}

		if err := plan.ChangeCoreDetails(companyName, startMonth, startYear); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.PlanSvc.Save(plan); err != nil {
			http.Error(w, "Failed to save plan", http.StatusInternalServerError)
			return
		}

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

		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
			return
		}

		invites, isOwner := app.invitesAndOwnerStatus(planID, user)

		page := views.BuildSetupPage(r, user, plan, invites, isOwner, toViewsHubCompletion(app.HubSvc.Get(planID)))
		views.RenderSetupPage(w, app.TemplateCache, page)
	}
}

// invitesAndOwnerStatus loads every invite for a plan and reports whether the
// current user is the plan's Owner.
func (app *App) invitesAndOwnerStatus(planID uuid.UUID, user *domain.User) ([]*domain.PlanInvite, bool) {
	invites, err := app.InviteSvc.GetInvitesForPlan(planID)
	if err != nil {
		log.Printf("Failed to load invites for plan %s: %v", planID, err)
		invites = nil
	}

	isOwner := false
	if user != nil {
		if access, err := app.AccessSvc.GetAccess(planID, user.ID()); err == nil {
			isOwner = access.AccessLevel == domain.Owner
		}
	}

	return invites, isOwner
}

func (app *App) GetIncomeStatement() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid Plan ID")
			return
		}

		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildIncomeStatementPage(r, GetUserFromContext(r), plan, toViewsHubCompletion(app.HubSvc.Get(planID)))
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

		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildBalanceSheetPage(r, GetUserFromContext(r), plan, toViewsHubCompletion(app.HubSvc.Get(planID)))
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

		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		page := views.BuildAnalyticsPage(r, GetUserFromContext(r), plan, toViewsHubCompletion(app.HubSvc.Get(planID)))
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

		if err := app.PlanSvc.Delete(id); err != nil {
			log.Printf("PlanSvc Delete Error: %v", err)
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
