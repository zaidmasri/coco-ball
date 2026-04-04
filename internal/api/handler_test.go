package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/store"
)

// setup sets up a fresh handler with a fresh memory store for each test.
func setup() (*Handler, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	handler := NewHandler(memStore)
	return handler, memStore
}

func TestHandler_CreatePlan(t *testing.T) {
	h, _ := setup()

	body, _ := json.Marshal(CreatePlanRequest{Name: "Test Plan", DurationMonths: 24})
	req := httptest.NewRequest("POST", "/plans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.handleCreatePlan(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	var resp PlanResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Name != "Test Plan" {
		t.Errorf("expected name 'Test Plan', got %s", resp.Name)
	}
	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
	if resp.DurationMonths != 24 {
		t.Errorf("expected DurationMonths 24, got %d", resp.DurationMonths)
	}
}

func TestHandler_GetPlans_Empty(t *testing.T) {
	h, _ := setup()

	req := httptest.NewRequest("GET", "/plans", nil)
	rr := httptest.NewRecorder()

	h.handleGetPlans(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "[]\n" {
		t.Errorf("expected empty array [], got %s", rr.Body.String())
	}
}

func TestHandler_AddExpense(t *testing.T) {
	h, _ := setup()

	createBody, _ := json.Marshal(CreatePlanRequest{Name: "Expense Plan", DurationMonths: 12})
	h.handleCreatePlan(httptest.NewRecorder(), httptest.NewRequest("POST", "/plans", bytes.NewBuffer(createBody)))

	// FIXED: Changed DurationMonths to MonthIndex
	expBody, _ := json.Marshal(FinancialEntryRequest{Name: "Rent", Amount: 1000, MonthIndex: 0})

	req := httptest.NewRequest("POST", "/plans/1/expenses", bytes.NewBuffer(expBody))
	req.SetPathValue("id", "1")

	rr := httptest.NewRecorder()
	h.handleAddExpense(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp PlanResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if len(resp.Expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(resp.Expenses))
	}
	if resp.Expenses[0].MonthIndex != 0 {
		t.Errorf("expected expense at MonthIndex 0, got %d", resp.Expenses[0].MonthIndex)
	}
}

func TestHandler_AddRevenue(t *testing.T) {
	h, _ := setup()

	createBody, _ := json.Marshal(CreatePlanRequest{Name: "Revenue Plan", DurationMonths: 12})
	h.handleCreatePlan(httptest.NewRecorder(), httptest.NewRequest("POST", "/plans", bytes.NewBuffer(createBody)))

	// FIXED: Changed DurationMonths to MonthIndex
	revBody, _ := json.Marshal(FinancialEntryRequest{Name: "Product Sales", Amount: 5000, MonthIndex: 11})

	req := httptest.NewRequest("POST", "/plans/1/revenues", bytes.NewBuffer(revBody))
	req.SetPathValue("id", "1")

	rr := httptest.NewRecorder()
	h.handleAddRevenue(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp PlanResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if len(resp.Revenues) != 1 {
		t.Fatalf("expected 1 revenue, got %d", len(resp.Revenues))
	}
	if resp.Revenues[0].MonthIndex != 11 {
		t.Errorf("expected revenue at MonthIndex 11, got %d", resp.Revenues[0].MonthIndex)
	}
}

func TestHandler_FinancialEntry_OutOfBounds(t *testing.T) {
	h, _ := setup()

	createBody, _ := json.Marshal(CreatePlanRequest{Name: "Strict Plan", DurationMonths: 12})
	h.handleCreatePlan(httptest.NewRecorder(), httptest.NewRequest("POST", "/plans", bytes.NewBuffer(createBody)))

	// FIXED: Changed DurationMonths to MonthIndex
	expBody, _ := json.Marshal(FinancialEntryRequest{Name: "Rent", Amount: 1000, MonthIndex: 12})

	req := httptest.NewRequest("POST", "/plans/1/expenses", bytes.NewBuffer(expBody))
	req.SetPathValue("id", "1")

	rr := httptest.NewRecorder()
	h.handleAddExpense(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for out of bounds month, got %d", rr.Code)
	}
}

func TestHandler_GetPlan_NotFound(t *testing.T) {
	h, _ := setup()

	req := httptest.NewRequest("GET", "/plans/999", nil)
	req.SetPathValue("id", "999")

	rr := httptest.NewRecorder()
	h.handleGetPlan(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}
