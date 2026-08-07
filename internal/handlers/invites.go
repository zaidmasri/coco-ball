package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// PostCreateInvite POST /plan/{id}/invites (Owner only - invites a
// collaborator to the plan at a chosen access level)
func (app *App) PostCreateInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		user := GetUserFromContext(r)

		renderError := func(errMsg string, statusCode int) {
			invites, isOwner := app.invitesAndOwnerStatus(planID, user)
			w.WriteHeader(statusCode)
			page := views.BuildSetupPage(r, user, plan, invites, isOwner, app.hubCompletion(planID))
			page.ErrorMessage = errMsg
			views.RenderSetupPage(w, app.TemplateCache, page)
		}

		email := strings.TrimSpace(r.PostForm.Get("email"))
		level := domain.AccessLevel(r.PostForm.Get("accessLevel"))

		if email == "" || !level.IsValid() {
			renderError("Please provide a valid email address and access level.", http.StatusBadRequest)
			return
		}

		invite, err := domain.NewPlanInvite(planID, email, level, user.ID())
		if err != nil {
			renderError(err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.InviteSvc.CreateInvite(invite); err != nil {
			log.Printf("Failed to create invite: %v", err)
			renderError("An internal database error occurred. Please try again.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plan/"+planID.String()+"/setup", http.StatusSeeOther)
	}
}

// PostAcceptInvite POST /invites/{id}/accept (grants the logged-in user
// access to the plan and redirects them to its first step)
func (app *App) PostAcceptInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			app.renderErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to respond to an invite.")
			return
		}

		invite, ok := app.loadOwnPendingInvite(w, r, user)
		if !ok {
			return
		}

		if err := app.AccessSvc.GrantAccess(invite.PlanID, user.ID(), invite.AccessLevel); err != nil {
			log.Printf("Failed to grant access from invite %s: %v", invite.ID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to accept invite. Please try again.")
			return
		}

		if err := app.InviteSvc.UpdateInviteStatus(invite.ID, domain.InviteAccepted); err != nil {
			log.Printf("Failed to update invite %s status: %v", invite.ID, err)
		}

		http.Redirect(w, r, "/plan/"+invite.PlanID.String()+"/starting-point", http.StatusSeeOther)
	}
}

// PostRejectInvite POST /invites/{id}/reject (records the decline and keeps
// it out of the invitee's onboarding section going forward)
func (app *App) PostRejectInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil {
			app.renderErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to respond to an invite.")
			return
		}

		invite, ok := app.loadOwnPendingInvite(w, r, user)
		if !ok {
			return
		}

		if err := app.InviteSvc.UpdateInviteStatus(invite.ID, domain.InviteRejected); err != nil {
			log.Printf("Failed to update invite %s status: %v", invite.ID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to decline invite. Please try again.")
			return
		}

		log.Printf("User %s declined invite %s to plan %s", user.Email(), invite.ID, invite.PlanID)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// loadOwnPendingInvite parses the invite ID from the request path and
// verifies it exists, is still pending, and is addressed to the logged-in
// user. On failure it renders an appropriate error page and returns ok=false.
func (app *App) loadOwnPendingInvite(w http.ResponseWriter, r *http.Request, user *domain.User) (*domain.PlanInvite, bool) {
	inviteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid invite ID")
		return nil, false
	}

	invite, err := app.InviteSvc.GetInvite(inviteID)
	if err != nil {
		app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that invite. It may have already been responded to.")
		return nil, false
	}

	if !strings.EqualFold(invite.Email, user.Email()) {
		app.renderErrorPage(w, r, http.StatusForbidden, "This invite was not sent to you.")
		return nil, false
	}

	if invite.Status != domain.InvitePending {
		app.renderErrorPage(w, r, http.StatusConflict, "This invite has already been responded to.")
		return nil, false
	}

	return invite, true
}
