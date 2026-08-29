package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// PlanController handles the plan lifecycle (creation, setup, deletion) and
// the top-level report pages (income statement, balance sheet, analytics)
// that sit outside any single wizard hub.
type PlanController struct {
	planSvc       ifaces.PlanService
	accessSvc     ifaces.AccessService
	inviteSvc     ifaces.InviteService
	authSvc       ifaces.AuthService
	hubSvc        ifaces.HubCompletionService
	templateCache map[string]*template.Template
}

func NewPlanController(
	mux *http.ServeMux,
	planSvc ifaces.PlanService,
	accessSvc ifaces.AccessService,
	inviteSvc ifaces.InviteService,
	authSvc ifaces.AuthService,
	hubSvc ifaces.HubCompletionService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *PlanController {
	c := &PlanController{
		planSvc:       planSvc,
		accessSvc:     accessSvc,
		inviteSvc:     inviteSvc,
		authSvc:       authSvc,
		hubSvc:        hubSvc,
		templateCache: templateCache,
	}

	mux.HandleFunc("GET /{$}", c.GetRoot)
	mux.HandleFunc("POST /plan/setup", c.PostSetup)
	mux.Handle("GET /plan/{id}/setup", accessMW.RequireAccess(domain.Viewer)(http.HandlerFunc(c.GetSetup)))
	mux.Handle("POST /plan/{id}/setup", accessMW.RequireAccess(domain.Editor)(http.HandlerFunc(c.PostUpdateSetup)))
	mux.Handle("GET /plan/{id}/income-statement", accessMW.RequireAccess(domain.Viewer)(http.HandlerFunc(c.GetIncomeStatement)))
	mux.Handle("GET /plan/{id}/balance-sheet", accessMW.RequireAccess(domain.Viewer)(http.HandlerFunc(c.GetBalanceSheet)))
	mux.Handle("GET /plan/{id}/analytics", accessMW.RequireAccess(domain.Viewer)(http.HandlerFunc(c.GetAnalytics)))
	mux.Handle("POST /plan/{id}/delete", accessMW.RequireAccess(domain.Owner)(http.HandlerFunc(c.PostDeletePlan)))

	return c
}

// toViewsHubCompletion translates the application-layer HubCompletion to the
// views-layer equivalent (same fields, different package).
func toViewsHubCompletion(c ifaces.HubCompletion) views.HubCompletion {
	return views.HubCompletion{
		StartingPoint:     c.StartingPoint,
		Payroll:           c.Payroll,
		SalesForecast:     c.SalesForecast,
		OperatingExpenses: c.OperatingExpenses,
		CashFlow:          c.CashFlow,
	}
}

// hubCompletion loads the completion state of all five wizard hubs, for
// pages whose sidebar needs every icon's fill state.
func hubCompletion(hubSvc ifaces.HubCompletionService, planID uuid.UUID) views.HubCompletion {
	return toViewsHubCompletion(hubSvc.Get(planID))
}

// invitesAndOwnerStatus loads every invite for a plan and reports whether
// the current user is the plan's Owner.
func invitesAndOwnerStatus(inviteSvc ifaces.InviteService, accessSvc ifaces.AccessService, planID uuid.UUID, user *domain.User) ([]*domain.PlanInvite, bool) {
	invites, err := inviteSvc.GetInvitesForPlan(planID)
	if err != nil {
		log.Printf("Failed to load invites for plan %s: %v", planID, err)
		invites = nil
	}

	isOwner := false
	if user != nil {
		if access, err := accessSvc.GetAccess(planID, user.ID()); err == nil {
			isOwner = access.AccessLevel == domain.Owner
		}
	}

	return invites, isOwner
}

func (c *PlanController) GetRoot(w http.ResponseWriter, r *http.Request) {
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
		views.RenderIndexPage(w, c.templateCache, page)
		return
	}

	plans, err := c.planSvc.GetUserPlans(user.ID())
	if err != nil {
		plans = []*domain.Plan{}
	}

	pendingInvites := c.pendingInvitesForUser(user)

	page := views.BuildIndexPage(r, user, plans, pendingInvites)
	views.RenderIndexPage(w, c.templateCache, page)
}

// pendingInvitesForUser loads pending invites for the user and enriches each
// with the plan name and inviter name for the landing page onboarding section.
func (c *PlanController) pendingInvitesForUser(user *domain.User) []views.InviteSummary {
	invites, err := c.inviteSvc.GetPendingInvitesForEmail(user.Email())
	if err != nil {
		log.Printf("Failed to load pending invites for %s: %v", user.Email(), err)
		return nil
	}

	summaries := make([]views.InviteSummary, 0, len(invites))
	for _, invite := range invites {
		summary := views.InviteSummary{Invite: invite}

		if plan, err := c.planSvc.Get(invite.PlanID); err == nil {
			summary.PlanName = plan.Name()
		} else {
			summary.PlanName = "Unknown Plan"
		}

		if inviter, err := c.authSvc.GetUser(invite.InvitedBy); err == nil {
			summary.InviterName = inviter.FullName()
		} else {
			summary.InviterName = "a collaborator"
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

// PostSetup POST /plan/setup (Saves a brand new plan)
func (c *PlanController) PostSetup(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		renderErrorPage(w, r, c.templateCache, http.StatusUnauthorized, "You must be logged in to create a plan.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
		return
	}

	companyName := r.PostForm.Get("companyName")
	startMonth, errMonth := strconv.Atoi(r.PostForm.Get("startMonth"))
	startYear, errYear := strconv.Atoi(r.PostForm.Get("startYear"))
	if errMonth != nil || errYear != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid month or year provided.")
		return
	}

	// CreatePlan grants the creator Owner access atomically, in the same
	// transaction as the plan row itself - no separate GrantAccess call
	// needed, so a grant failure can never leave an ownerless plan behind.
	result, err := c.planSvc.CreatePlan(&commands.CreatePlan{
		Name:          companyName,
		StartingMonth: startMonth,
		StartingYear:  startYear,
		OwnerID:       user.ID(),
	})
	if err != nil {
		renderCommandError(w, r, c.templateCache, "PlanSvc CreatePlan", err)
		return
	}
	planID := result.Result.ID

	http.Redirect(w, r, "/plan/"+planID.String()+"/starting-point", http.StatusSeeOther)
}

// PostUpdateSetup POST /plan/{id}/setup (Updates an existing plan)
func (c *PlanController) PostUpdateSetup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid Plan ID")
		return
	}

	// Existence check up front so a missing plan 404s rather than surfacing
	// as a 400 from UpdatePlan's internal Get.
	if _, err := c.planSvc.Get(id); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	companyName := r.PostForm.Get("companyName")
	startMonth, errMonth := strconv.Atoi(r.PostForm.Get("startMonth"))
	startYear, errYear := strconv.Atoi(r.PostForm.Get("startYear"))

	if errMonth != nil || errYear != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid month or year provided.")
		return
	}

	if _, err := c.planSvc.UpdatePlan(&commands.UpdatePlan{
		PlanID:        id,
		Name:          companyName,
		StartingMonth: startMonth,
		StartingYear:  startYear,
	}); err != nil {
		renderCommandError(w, r, c.templateCache, "PlanSvc UpdatePlan", err)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

// GetSetup GET /plan/{id}/setup (Loads an existing plan into the form)
func (c *PlanController) GetSetup(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
		return
	}

	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
		return
	}

	invites, isOwner := invitesAndOwnerStatus(c.inviteSvc, c.accessSvc, planID, user)

	page := views.BuildSetupPage(r, user, plan, invites, isOwner, hubCompletion(c.hubSvc, planID))
	views.RenderSetupPage(w, c.templateCache, page)
}

func (c *PlanController) GetIncomeStatement(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid Plan ID")
		return
	}

	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	page := views.BuildIncomeStatementPage(r, GetUserFromContext(r), plan, hubCompletion(c.hubSvc, planID))
	views.RenderIncomeStatementPage(w, c.templateCache, page)
}

func (c *PlanController) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid Plan ID")
		return
	}

	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	page := views.BuildBalanceSheetPage(r, GetUserFromContext(r), plan, hubCompletion(c.hubSvc, planID))
	views.RenderBalanceSheetPage(w, c.templateCache, page)
}

func (c *PlanController) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid Plan ID")
		return
	}

	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	page := views.BuildAnalyticsPage(r, GetUserFromContext(r), plan, hubCompletion(c.hubSvc, planID))
	views.RenderAnalyticsPage(w, c.templateCache, page)
}

// PostDeletePlan POST /plan/{id}/delete (requires Owner access)
func (c *PlanController) PostDeletePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	if _, err := c.planSvc.DeletePlan(&commands.DeletePlan{PlanID: id}); err != nil {
		renderInternalError(w, r, c.templateCache, "PlanSvc DeletePlan", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
