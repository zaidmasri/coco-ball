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

	// 1. Prepare the JSON request body
	body, _ := json.Marshal(CreatePlanRequest{Name: "Test Plan"})

	// 2. Create a fake POST request
	req := httptest.NewRequest("POST", "/plans", bytes.NewBuffer(body))

	// 3. Create a recorder to capture the response
	rr := httptest.NewRecorder()

	// 4. Call the handler directly
	h.handleCreatePlan(rr, req)

	// 5. Assertions
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	var resp PlanResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Name != "Test Plan" {
		t.Errorf("expected name 'Test Plan', got %s", resp.Name)
	}
	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
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

	// Check if it returns an empty array [] instead of null
	if rr.Body.String() != "[]\n" {
		t.Errorf("expected empty array [], got %s", rr.Body.String())
	}
}

func TestHandler_AddExpense(t *testing.T) {
	h, _ := setup()

	// Pre-populate a plan so we can add an expense to it
	// planID := s.GenerateID()
	// (Note: In a real test you'd use the domain constructor,
	// but since we're testing the API flow, we'll just prep the store)
	// For simplicity in this example, we assume ID 1 exists after a create.

	// We'll simulate a POST to /plans first
	createBody, _ := json.Marshal(CreatePlanRequest{Name: "Expense Plan"})
	h.handleCreatePlan(httptest.NewRecorder(), httptest.NewRequest("POST", "/plans", bytes.NewBuffer(createBody)))

	// Now add the expense
	expBody, _ := json.Marshal(AddExpenseRequest{Name: "Rent", Amount: 1000})

	// Go 1.22+ PathValue requires a bit of manual setup in httptest
	// unless you wrap the handler in a mux.
	req := httptest.NewRequest("POST", "/plans/1/expenses", bytes.NewBuffer(expBody))
	req.SetPathValue("id", "1") // Manually set the path value for the test

	rr := httptest.NewRecorder()
	h.handleAddExpense(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp PlanResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if len(resp.Expenses) != 1 {
		t.Errorf("expected 1 expense, got %d", len(resp.Expenses))
	}
	if resp.TotalExpenses != 1000 {
		t.Errorf("expected total 1000, got %d", resp.TotalExpenses)
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
