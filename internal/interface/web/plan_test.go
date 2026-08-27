package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// TestReportPagesAccessControl exercises the plan-access gate shared by the
// three report pages (Income Statement, Balance Sheet, Analytics), each
// registered behind accessMW.RequireAccess(domain.Viewer) in plan.go.
func TestReportPagesAccessControl(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "owner@example.com")
	outsider := createTestUser(t, s, "outsider@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)

	reportRoutes := []string{
		"/plan/" + plan.ID().String() + "/income-statement",
		"/plan/" + plan.ID().String() + "/balance-sheet",
		"/plan/" + plan.ID().String() + "/analytics",
	}

	for _, route := range reportRoutes {
		t.Run(route+" unauthenticated", func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", w.Code)
			}
		})

		t.Run(route+" no access", func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			req = WithUserContext(req, outsider)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Errorf("expected 403, got %d", w.Code)
			}
		})

		t.Run(route+" owner can view", func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			req = WithUserContext(req, owner)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestReportPagesRenderForEachAccessLevel confirms that Owner, Editor, and
// Viewer can all load each report page (they only require Viewer-level
// access) and that the real page template renders without error - this is
// the case a fully mocked template cache can't catch, since a missing field
// or bad template reference only surfaces once html/template actually
// executes the real templates.
func TestReportPagesRenderForEachAccessLevel(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "owner2@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)

	levels := []struct {
		name  string
		email string
		level domain.AccessLevel
	}{
		{"Owner", "owner2@example.com", domain.Owner},
		{"Editor", "editor2@example.com", domain.Editor},
		{"Viewer", "viewer2@example.com", domain.Viewer},
	}

	pages := []struct {
		name    string
		route   string
		heading string
	}{
		{"income statement", "/plan/" + plan.ID().String() + "/income-statement", "Income Statement (P&L)"},
		{"balance sheet", "/plan/" + plan.ID().String() + "/balance-sheet", "Balance Sheet"},
		{"analytics", "/plan/" + plan.ID().String() + "/analytics", "Analytics & Diagnostics"},
	}

	for _, lvl := range levels {
		var user *domain.User
		if lvl.name == "Owner" {
			user = owner
		} else {
			user = createTestUser(t, s, lvl.email)
			if err := s.access.GrantAccess(plan.ID(), user.ID(), lvl.level); err != nil {
				t.Fatalf("failed to grant %s access: %v", lvl.name, err)
			}
		}

		for _, pg := range pages {
			t.Run(lvl.name+" views "+pg.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", pg.route, nil)
				req = WithUserContext(req, user)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
				}
				if !strings.Contains(w.Body.String(), pg.heading) {
					t.Errorf("expected rendered page to contain heading %q", pg.heading)
				}
			})
		}
	}
}
