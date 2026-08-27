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

// newInventoryPurchaseDraft POSTs the "new" route and returns the created
// item's ID, extracted from the redirect Location
// (/plan/{planID}/cash-flow/inventory-purchases/{itemID}/category).
func newInventoryPurchaseDraft(t *testing.T, mux *http.ServeMux, planID string, user *domain.User) string {
	t.Helper()
	w := postForm(mux, "/plan/"+planID+"/cash-flow/inventory-purchases/new", user, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 creating inventory purchase draft, got %d: %s", w.Code, w.Body.String())
	}
	parts := strings.Split(strings.Trim(w.Header().Get("Location"), "/"), "/")
	if len(parts) < 2 {
		t.Fatalf("unexpected redirect location %q", w.Header().Get("Location"))
	}
	return parts[len(parts)-2]
}

func TestCashFlowInventoryPurchaseHappyPath(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "cf-owner@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newInventoryPurchaseDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/cash-flow/inventory-purchases/" + itemID + "/"

	steps := []struct {
		step string
		form url.Values
	}{
		{"category", url.Values{"name": {"Raw Materials"}}},
		{"monthly-amount", url.Values{"monthly_amount": {"2000"}}},
		{"growth-yr2", url.Values{"growth_yr2": {"4"}}},
		{"growth-yr3", url.Values{"growth_yr3": {"2"}}},
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
	item, err := s.inventoryPurchase.GetInventoryPurchase(id)
	if err != nil {
		t.Fatalf("failed to load saved inventory purchase: %v", err)
	}
	if item.Status != repositories.StatusComplete {
		t.Errorf("expected status Complete, got %v", item.Status)
	}
	if item.Purchase.Category != "Raw Materials" {
		t.Errorf("expected category 'Raw Materials', got %q", item.Purchase.Category)
	}
	if item.Purchase.MonthlyAmount.MinorUnits() != 2000 {
		t.Errorf("expected monthly amount 2000, got %d", item.Purchase.MonthlyAmount.MinorUnits())
	}
}

func TestCashFlowInventoryPurchaseAuthorization(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "cf-owner2@example.com")
	viewer := createTestUser(t, s, "cf-viewer@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	if err := s.access.GrantAccess(plan.ID(), viewer.ID(), domain.Viewer); err != nil {
		t.Fatalf("failed to grant viewer access: %v", err)
	}

	w := postForm(mux, "/plan/"+plan.ID().String()+"/cash-flow/inventory-purchases/new", viewer, url.Values{})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer creating an inventory purchase, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCashFlowInventoryPurchaseValidation(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "cf-owner3@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newInventoryPurchaseDraft(t, mux, planID, owner)

	w := postForm(mux, "/plan/"+planID+"/cash-flow/inventory-purchases/"+itemID+"/category", owner, url.Values{"name": {""}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for blank category, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCashFlowInventoryPurchaseDelete(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "cf-owner4@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newInventoryPurchaseDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/cash-flow/inventory-purchases/" + itemID + "/"
	steps := []struct {
		step string
		form url.Values
	}{
		{"category", url.Values{"name": {"Packaging"}}},
		{"monthly-amount", url.Values{"monthly_amount": {"500"}}},
		{"growth-yr2", url.Values{}},
		{"growth-yr3", url.Values{}},
	}
	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			t.Fatalf("step %s: expected 303 or 200, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	w := postForm(mux, "/plan/"+planID+"/cash-flow/inventory-purchases/"+itemID, owner, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 deleting inventory purchase, got %d: %s", w.Code, w.Body.String())
	}

	purchases, err := s.inventoryPurchase.ListCompleteInventoryPurchases(plan.ID())
	if err != nil {
		t.Fatalf("failed to list inventory purchases: %v", err)
	}
	for _, p := range purchases {
		if p.ID.String() == itemID {
			t.Error("expected deleted inventory purchase to be excluded from ListCompleteInventoryPurchases")
		}
	}
}
