package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// postForm submits a POST with the given form values as the logged-in user
// (or unauthenticated if user is nil) and returns the recorded response.
func postForm(mux *http.ServeMux, path string, user *domain.User, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, nil)
	req.PostForm = form
	if user != nil {
		req = WithUserContext(req, user)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func TestPostCreateInvite(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "owner@example.com")
	editor := createTestUser(t, s, "editor@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	if err := s.access.GrantAccess(plan.ID(), editor.ID(), domain.Editor); err != nil {
		t.Fatalf("failed to grant editor access: %v", err)
	}

	t.Run("owner can invite a collaborator", func(t *testing.T) {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {"collaborator@example.com"},
			"accessLevel": {string(domain.Viewer)},
		})
		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
		}

		invites, err := s.invites.GetPendingInvitesForEmail("collaborator@example.com")
		if err != nil {
			t.Fatalf("failed to load invites: %v", err)
		}
		if len(invites) != 1 {
			t.Fatalf("expected 1 pending invite, got %d", len(invites))
		}
		if invites[0].AccessLevel != domain.Viewer {
			t.Errorf("expected Viewer access level, got %s", invites[0].AccessLevel)
		}
	})

	t.Run("editor cannot invite a collaborator", func(t *testing.T) {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", editor, url.Values{
			"email":       {"another@example.com"},
			"accessLevel": {string(domain.Viewer)},
		})
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing email is rejected", func(t *testing.T) {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {""},
			"accessLevel": {string(domain.Viewer)},
		})
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("owner cannot invite themselves", func(t *testing.T) {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {"owner@example.com"},
			"accessLevel": {string(domain.Viewer)},
		})
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
		invites, err := s.invites.GetPendingInvitesForEmail("owner@example.com")
		if err != nil {
			t.Fatalf("failed to load invites: %v", err)
		}
		if len(invites) != 0 {
			t.Error("expected no invite to be created for a self-invite")
		}
	})

	t.Run("duplicate pending invite is rejected", func(t *testing.T) {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {"repeat@example.com"},
			"accessLevel": {string(domain.Viewer)},
		})
		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected first invite to succeed with 303, got %d: %s", w.Code, w.Body.String())
		}

		w = postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {"repeat@example.com"},
			"accessLevel": {string(domain.Editor)},
		})
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for duplicate invite, got %d: %s", w.Code, w.Body.String())
		}

		invites, err := s.invites.GetPendingInvitesForEmail("repeat@example.com")
		if err != nil {
			t.Fatalf("failed to load invites: %v", err)
		}
		if len(invites) != 1 {
			t.Errorf("expected exactly 1 pending invite, got %d", len(invites))
		}
	})
}

func TestPostAcceptInvite(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "owner2@example.com")
	invitee := createTestUser(t, s, "invitee@example.com")
	otherUser := createTestUser(t, s, "notinvited@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)

	createInvite := func(email string, level domain.AccessLevel) *domain.PlanInvite {
		w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
			"email":       {email},
			"accessLevel": {string(level)},
		})
		if w.Code != http.StatusSeeOther {
			t.Fatalf("failed to create invite for %s: %d %s", email, w.Code, w.Body.String())
		}
		invites, err := s.invites.GetPendingInvitesForEmail(email)
		if err != nil || len(invites) == 0 {
			t.Fatalf("failed to fetch created invite for %s: %v", email, err)
		}
		return invites[0]
	}

	t.Run("invited user can accept and is granted access", func(t *testing.T) {
		invite := createInvite("invitee@example.com", domain.Editor)

		w := postForm(mux, "/invites/"+invite.ID.String()+"/accept", invitee, nil)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
		}

		access, err := s.access.GetAccess(plan.ID(), invitee.ID())
		if err != nil {
			t.Fatalf("expected access to be granted: %v", err)
		}
		if access.AccessLevel != domain.Editor {
			t.Errorf("expected Editor access, got %s", access.AccessLevel)
		}
	})

	t.Run("a user the invite was not addressed to is forbidden", func(t *testing.T) {
		invite := createInvite("someone-else@example.com", domain.Viewer)

		w := postForm(mux, "/invites/"+invite.ID.String()+"/accept", otherUser, nil)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
		}

		if _, err := s.access.GetAccess(plan.ID(), otherUser.ID()); err == nil {
			t.Error("expected no access to be granted to the wrong user")
		}
	})

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		invite := createInvite("unauth-flow@example.com", domain.Viewer)

		w := postForm(mux, "/invites/"+invite.ID.String()+"/accept", nil, nil)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestPostRejectInvite(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "owner3@example.com")
	invitee := createTestUser(t, s, "rejecting@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)

	w := postForm(mux, "/plan/"+plan.ID().String()+"/invites", owner, url.Values{
		"email":       {"rejecting@example.com"},
		"accessLevel": {string(domain.Viewer)},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("failed to create invite: %d %s", w.Code, w.Body.String())
	}
	invites, err := s.invites.GetPendingInvitesForEmail("rejecting@example.com")
	if err != nil || len(invites) == 0 {
		t.Fatalf("failed to fetch created invite: %v", err)
	}
	invite := invites[0]

	w = postForm(mux, "/invites/"+invite.ID.String()+"/reject", invitee, nil)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := s.access.GetAccess(plan.ID(), invitee.ID()); err == nil {
		t.Error("expected no access to be granted after rejecting an invite")
	}

	updated, err := s.invites.GetInvite(invite.ID)
	if err != nil {
		t.Fatalf("failed to reload invite: %v", err)
	}
	if updated.Status != domain.InviteRejected {
		t.Errorf("expected invite status Rejected, got %s", updated.Status)
	}
}
