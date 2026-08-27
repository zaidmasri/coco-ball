package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// newSalaryRoleDraft POSTs the "new" route and returns the created item's
// ID, extracted from the redirect Location the same way a browser would
// follow it (Location is /plan/{planID}/payroll/salary-roles/{itemID}/role).
func newSalaryRoleDraft(t *testing.T, mux *http.ServeMux, planID string, user *domain.User) string {
	t.Helper()
	w := postForm(mux, "/plan/"+planID+"/payroll/salary-roles/new", user, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 creating salary role draft, got %d: %s", w.Code, w.Body.String())
	}
	parts := strings.Split(strings.Trim(w.Header().Get("Location"), "/"), "/")
	if len(parts) < 2 {
		t.Fatalf("unexpected redirect location %q", w.Header().Get("Location"))
	}
	return parts[len(parts)-2]
}

func TestPayrollSalaryRoleHappyPath(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "payroll-owner@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newSalaryRoleDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/payroll/salary-roles/" + itemID + "/"

	steps := []struct {
		step string
		form url.Values
	}{
		{"role", url.Values{"name": {"Software Engineer"}}},
		{"type", url.Values{"is_contractor": {"false"}}},
		{"headcount", url.Values{"headcount": {"3"}}},
		{"monthly-pay", url.Values{"monthly_pay": {"9000"}}},
		{"growth-yr2", url.Values{"growth_yr2": {"5"}}},
		{"growth-yr3", url.Values{"growth_yr3": {"3"}}},
	}

	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			t.Fatalf("step %s: expected 303 or 200, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	id, err := uuid.Parse(itemID)
	if err != nil {
		t.Fatalf("failed to parse item ID: %v", err)
	}
	item, err := s.salaryRoles.GetSalaryRole(id)
	if err != nil {
		t.Fatalf("failed to load saved salary role: %v", err)
	}
	if item.Status != repositories.StatusComplete {
		t.Errorf("expected status Complete, got %v", item.Status)
	}
	if item.Role.Role != "Software Engineer" {
		t.Errorf("expected role name 'Software Engineer', got %q", item.Role.Role)
	}
	if item.Role.Headcount != 3 {
		t.Errorf("expected headcount 3, got %d", item.Role.Headcount)
	}
	if item.Role.MonthlyPay.MinorUnits() != 9000 {
		t.Errorf("expected monthly pay of 9000, got %d", item.Role.MonthlyPay.MinorUnits())
	}
}

func TestPayrollSalaryRoleAuthorization(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "payroll-owner2@example.com")
	viewer := createTestUser(t, s, "payroll-viewer@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	if err := s.access.GrantAccess(plan.ID(), viewer.ID(), domain.Viewer); err != nil {
		t.Fatalf("failed to grant viewer access: %v", err)
	}

	w := postForm(mux, "/plan/"+plan.ID().String()+"/payroll/salary-roles/new", viewer, url.Values{})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer creating a salary role, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPayrollSalaryRoleValidation(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "payroll-owner3@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newSalaryRoleDraft(t, mux, planID, owner)

	w := postForm(mux, "/plan/"+planID+"/payroll/salary-roles/"+itemID+"/role", owner, url.Values{"name": {""}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for blank role name, got %d: %s", w.Code, w.Body.String())
	}

	id, err := uuid.Parse(itemID)
	if err != nil {
		t.Fatalf("failed to parse item ID: %v", err)
	}
	item, err := s.salaryRoles.GetSalaryRole(id)
	if err != nil {
		t.Fatalf("failed to load salary role: %v", err)
	}
	if item.Status == repositories.StatusComplete {
		t.Error("expected draft to remain incomplete after validation failure")
	}
}

func TestPayrollSalaryRoleRejectsOutOfRangeGrowthRate(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "payroll-owner5@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newSalaryRoleDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/payroll/salary-roles/" + itemID + "/"

	steps := []struct {
		step string
		form url.Values
	}{
		{"role", url.Values{"name": {"Founder"}}},
		{"type", url.Values{"is_contractor": {"false"}}},
		{"headcount", url.Values{"headcount": {"1"}}},
		{"monthly-pay", url.Values{"monthly_pay": {"5000"}}},
	}
	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("step %s: expected 303, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	// growth-yr2 isn't the finishing step, so an out-of-range rate here is
	// only saved as an unvalidated draft (SaveSalaryRoleStep only runs
	// domain.ValidateSalaryRole on StatusComplete) - it surfaces once the
	// wizard reaches growth-yr3 and the full role (both rates) is validated.
	w := postForm(mux, base+"growth-yr2", owner, url.Values{"growth_yr2": {"50000"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 saving an intermediate draft step, got %d: %s", w.Code, w.Body.String())
	}

	w = postForm(mux, base+"growth-yr3", owner, url.Values{"growth_yr3": {"3"}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 finishing with a 50000%% Year 2 growth rate, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPayrollSalaryRoleDelete(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "payroll-owner4@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newSalaryRoleDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/payroll/salary-roles/" + itemID + "/"
	steps := []struct {
		step string
		form url.Values
	}{
		{"role", url.Values{"name": {"Designer"}}},
		{"type", url.Values{"is_contractor": {"false"}}},
		{"headcount", url.Values{"headcount": {"1"}}},
		{"monthly-pay", url.Values{"monthly_pay": {"5000"}}},
		{"growth-yr2", url.Values{}},
		{"growth-yr3", url.Values{}},
	}
	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			t.Fatalf("step %s: expected 303 or 200, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	w := postForm(mux, "/plan/"+planID+"/payroll/salary-roles/"+itemID, owner, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 deleting salary role, got %d: %s", w.Code, w.Body.String())
	}

	roles, err := s.salaryRoles.ListCompleteSalaryRoles(plan.ID())
	if err != nil {
		t.Fatalf("failed to list salary roles: %v", err)
	}
	for _, r := range roles {
		if r.ID.String() == itemID {
			t.Error("expected deleted salary role to be excluded from ListCompleteSalaryRoles")
		}
	}
}
