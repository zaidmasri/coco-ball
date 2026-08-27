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

// newProductDraft POSTs the "new" route and returns the created item's ID,
// extracted from the redirect Location (/plan/{planID}/sales-forecast/products/{itemID}/name).
func newProductDraft(t *testing.T, mux *http.ServeMux, planID string, user *domain.User) string {
	t.Helper()
	w := postForm(mux, "/plan/"+planID+"/sales-forecast/products/new", user, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 creating product draft, got %d: %s", w.Code, w.Body.String())
	}
	parts := strings.Split(strings.Trim(w.Header().Get("Location"), "/"), "/")
	if len(parts) < 2 {
		t.Fatalf("unexpected redirect location %q", w.Header().Get("Location"))
	}
	return parts[len(parts)-2]
}

func TestSalesForecastProductHappyPath(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "sf-owner@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newProductDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/sales-forecast/products/" + itemID + "/"

	steps := []struct {
		step string
		form url.Values
	}{
		{"name", url.Values{"name": {"Widget"}}},
		{"month1-units", url.Values{"month1_units": {"100"}}},
		{"price", url.Values{"price": {"25"}}},
		{"cost", url.Values{"cost": {"10"}}},
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
	item, err := s.products.GetProduct(id)
	if err != nil {
		t.Fatalf("failed to load saved product: %v", err)
	}
	if item.Status != repositories.StatusComplete {
		t.Errorf("expected status Complete, got %v", item.Status)
	}
	if item.Product.Name != "Widget" {
		t.Errorf("expected product name 'Widget', got %q", item.Product.Name)
	}
	if item.Product.Month1Units != 100 {
		t.Errorf("expected 100 month-1 units, got %d", item.Product.Month1Units)
	}
	if item.Product.PricePerUnit.MinorUnits() != 25 {
		t.Errorf("expected price 25, got %d", item.Product.PricePerUnit.MinorUnits())
	}
}

func TestSalesForecastProductAuthorization(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "sf-owner2@example.com")
	viewer := createTestUser(t, s, "sf-viewer@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	if err := s.access.GrantAccess(plan.ID(), viewer.ID(), domain.Viewer); err != nil {
		t.Fatalf("failed to grant viewer access: %v", err)
	}

	w := postForm(mux, "/plan/"+plan.ID().String()+"/sales-forecast/products/new", viewer, url.Values{})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer creating a product, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSalesForecastProductValidation(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "sf-owner3@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newProductDraft(t, mux, planID, owner)

	w := postForm(mux, "/plan/"+planID+"/sales-forecast/products/"+itemID+"/name", owner, url.Values{"name": {""}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for blank product name, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSalesForecastProductDelete(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	owner := createTestUser(t, s, "sf-owner4@example.com")
	plan := createTestPlan(t, s, owner, domain.Owner)
	planID := plan.ID().String()

	itemID := newProductDraft(t, mux, planID, owner)
	base := "/plan/" + planID + "/sales-forecast/products/" + itemID + "/"
	steps := []struct {
		step string
		form url.Values
	}{
		{"name", url.Values{"name": {"Gadget"}}},
		{"month1-units", url.Values{"month1_units": {"50"}}},
		{"price", url.Values{"price": {"15"}}},
		{"cost", url.Values{"cost": {"5"}}},
	}
	for _, st := range steps {
		w := postForm(mux, base+st.step, owner, st.form)
		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			t.Fatalf("step %s: expected 303 or 200, got %d: %s", st.step, w.Code, w.Body.String())
		}
	}

	w := postForm(mux, "/plan/"+planID+"/sales-forecast/products/"+itemID, owner, url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 deleting product, got %d: %s", w.Code, w.Body.String())
	}

	products, err := s.products.ListCompleteProducts(plan.ID())
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	for _, p := range products {
		if p.ID.String() == itemID {
			t.Error("expected deleted product to be excluded from ListCompleteProducts")
		}
	}
}
