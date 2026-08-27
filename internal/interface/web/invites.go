package web

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// InviteController handles inviting collaborators to a plan and responding
// to those invites.
type InviteController struct {
	inviteSvc     ifaces.InviteService
	planSvc       ifaces.PlanService
	accessSvc     ifaces.AccessService
	hubSvc        ifaces.HubCompletionService
	templateCache map[string]*template.Template
}

func NewInviteController(
	mux *http.ServeMux,
	inviteSvc ifaces.InviteService,
	planSvc ifaces.PlanService,
	accessSvc ifaces.AccessService,
	hubSvc ifaces.HubCompletionService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *InviteController {
	c := &InviteController{
		inviteSvc:     inviteSvc,
		planSvc:       planSvc,
		accessSvc:     accessSvc,
		hubSvc:        hubSvc,
		templateCache: templateCache,
	}

	// Invites - only the plan owner can invite collaborators
	mux.Handle("POST /plan/{id}/invites", accessMW.RequireAccess(domain.Owner)(http.HandlerFunc(c.PostCreateInvite)))
	// Accept/reject are keyed by invite ID, not plan ID, so auth is checked
	// inside the handler rather than via PlanAccessMiddleware.
	mux.HandleFunc("POST /invites/{id}/accept", c.PostAcceptInvite)
	mux.HandleFunc("POST /invites/{id}/reject", c.PostRejectInvite)

	return c
}

// PostCreateInvite POST /plan/{id}/invites (Owner only - invites a
// collaborator to the plan at a chosen access level)
func (c *InviteController) PostCreateInvite(w http.ResponseWriter, r *http.Request) {
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

	user := GetUserFromContext(r)

	renderError := func(errMsg string, statusCode int) {
		invites, isOwner := invitesAndOwnerStatus(c.inviteSvc, c.accessSvc, planID, user)
		w.WriteHeader(statusCode)
		page := views.BuildSetupPage(r, user, plan, invites, isOwner, hubCompletion(c.hubSvc, planID))
		page.ErrorMessage = errMsg
		views.RenderSetupPage(w, c.templateCache, page)
	}

	email := strings.TrimSpace(r.PostForm.Get("email"))
	level := domain.AccessLevel(r.PostForm.Get("accessLevel"))

	if email == "" || !level.IsValid() {
		renderError("Please provide a valid email address and access level.", http.StatusBadRequest)
		return
	}

	// InviteService.CreateInvite constructs the PlanInvite (via
	// domain.NewPlanInvite) and persists it atomically with the Plan
	// aggregate's UserInvitedToPlan outbox event - no separate
	// construct-then-save-the-plan steps here, so a failure partway through
	// can never silently drop the event or leave an orphaned invite row.
	if _, err := c.inviteSvc.CreateInvite(&commands.CreateInvite{
		PlanID:       planID,
		Email:        email,
		InviterEmail: user.Email(),
		AccessLevel:  level,
		InvitedBy:    user.ID(),
	}); err != nil {
		renderError(err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/plan/"+planID.String()+"/setup", http.StatusSeeOther)
}

// PostAcceptInvite POST /invites/{id}/accept (grants the logged-in user
// access to the plan and redirects them to its first step)
func (c *InviteController) PostAcceptInvite(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		renderErrorPage(w, r, c.templateCache, http.StatusUnauthorized, "You must be logged in to respond to an invite.")
		return
	}

	invite, ok := c.loadOwnPendingInvite(w, r, user)
	if !ok {
		return
	}

	if err := c.accessSvc.GrantAccess(invite.PlanID, user.ID(), invite.AccessLevel); err != nil {
		log.Printf("Failed to grant access from invite %s: %v", invite.ID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to accept invite. Please try again.")
		return
	}

	if err := c.inviteSvc.UpdateInviteStatus(invite.ID, domain.InviteAccepted); err != nil {
		log.Printf("Failed to update invite %s status: %v", invite.ID, err)
	}

	http.Redirect(w, r, "/plan/"+invite.PlanID.String()+"/starting-point", http.StatusSeeOther)
}

// PostRejectInvite POST /invites/{id}/reject (records the decline and keeps
// it out of the invitee's onboarding section going forward)
func (c *InviteController) PostRejectInvite(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		renderErrorPage(w, r, c.templateCache, http.StatusUnauthorized, "You must be logged in to respond to an invite.")
		return
	}

	invite, ok := c.loadOwnPendingInvite(w, r, user)
	if !ok {
		return
	}

	if err := c.inviteSvc.UpdateInviteStatus(invite.ID, domain.InviteRejected); err != nil {
		log.Printf("Failed to update invite %s status: %v", invite.ID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to decline invite. Please try again.")
		return
	}

	log.Printf("User %s declined invite %s to plan %s", user.Email(), invite.ID, invite.PlanID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// loadOwnPendingInvite parses the invite ID from the request path and
// verifies it exists, is still pending, and is addressed to the logged-in
// user. On failure it renders an appropriate error page and returns ok=false.
func (c *InviteController) loadOwnPendingInvite(w http.ResponseWriter, r *http.Request, user *domain.User) (*domain.PlanInvite, bool) {
	inviteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid invite ID")
		return nil, false
	}

	invite, err := c.inviteSvc.GetInvite(inviteID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that invite. It may have already been responded to.")
		return nil, false
	}

	if !strings.EqualFold(invite.Email, user.Email()) {
		renderErrorPage(w, r, c.templateCache, http.StatusForbidden, "This invite was not sent to you.")
		return nil, false
	}

	if invite.Status != domain.InvitePending {
		renderErrorPage(w, r, c.templateCache, http.StatusConflict, "This invite has already been responded to.")
		return nil, false
	}

	return invite, true
}
