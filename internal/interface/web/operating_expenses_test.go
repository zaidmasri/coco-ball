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

// newOperatingExpenseDraft POSTs the "new" route and returns the created
// item's ID, extracted from the redirect Location
// (/plan/{planID}/operating-expenses/{itemID}/name).
func newOperatingExpenseDraft(t *testing.T, mux *http.ServeMux, planID string, user *domain.User) string {
	t.Helper()
	w := postForm(mux, "/plan/"+planID+"/operating-expenses/new", user, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 creating operating expense draft, got %d: %s", w.Code, w.Body.String())
	}
	parts := strings.Split(strings.Trim(w.Header().Get("Location"), "/"), "/")
	if len(parts) < 2 {
		t.Fatalf("unexpected redirect location %q", w.Header().Get("Location"))
	}
	return parts[len(parts)-2]
}

func TestOperatingExpenseHappyPath(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "opex-owner@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newOperatingExpenseDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/operating-expenses/" + itemID + "/"

	steps := []struct {
		step string
		form url.Values
	}{
		{"name", url.Values{"name": {"Rent"}}},
		{"amount", url.Values{"amount": {"3000"}}},
		{"growth", url.Values{"growth": {"2"}}},
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
	item, err := s.operatingExpenses.GetOperatingExpense(id)
	if err != nil {
		t.Fatalf("failed to load saved operating expense: %v", err)
	}
	if item.Status != repositories.StatusComplete {
		t.Errorf("expected status Complete, got %v", item.Status)
	}
	if item.Cost.Name() != "Rent" {
		t.Errorf("expected name 'Rent', got %q", item.Cost.Name())
	}
	if item.Cost.BaseAmountPerMonth().MinorUnits() != 3000 {
		t.Errorf("expected base amount 3000, got %d", item.Cost.BaseAmountPerMonth().MinorUnits())
	}
}

func TestOperatingExpenseAuthorization(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "opex-owner2@example.com")
	viewer := createTestUser(t, s, "opex-viewer@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	if err := s.access.GrantAccess(plan.ID(), viewer.ID(), domain.Viewer); err != nil {
		t.Fatalf("failed to grant viewer access: %v", err)
	}

	w := postForm(mux, "/plan/"+plan.ID().String()+"/operating-expenses/new", viewer, url.Values{})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer creating an operating expense, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOperatingExpenseValidation(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "opex-owner3@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newOperatingExpenseDraft(t, mux, planID, owner)

	w := postForm(mux, "/plan/"+planID+"/operating-expenses/"+itemID+"/name", owner, url.Values{"name": {""}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for blank expense name, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOperatingExpenseDelete(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "opex-owner4@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newOperatingExpenseDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/operating-expenses/" + itemID + "/"
	steps := []struct {
		step string
		form url.Values
	}{
		{"name", url.Values{"name": {"Utilities"}}},
		{"amount", url.Values{"amount": {"400"}}},
		{"growth", url.Values{"growth": {"0"}}},
	}
	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			t.Fatalf("step %s: expected 303 or 200, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	w := postForm(mux, "/plan/"+planID+"/operating-expenses/"+itemID, owner, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 deleting operating expense, got %d: %s", w.Code, w.Body.String())
	}

	expenses, err := s.operatingExpenses.ListCompleteOperatingExpenses(plan.ID())
	if err != nil {
		t.Fatalf("failed to list operating expenses: %v", err)
	}
	for _, e := range expenses {
		if e.ID.String() == itemID {
			t.Error("expected deleted operating expense to be excluded from ListCompleteOperatingExpenses")
		}
	}
}
