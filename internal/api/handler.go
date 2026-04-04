package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/store"
)

type Handler struct {
	store store.PlanStore // Depends on the interface, not the concrete implementation
}

func NewHandler(s store.PlanStore) *Handler {
	return &Handler{store: s}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans", h.handleCreatePlan)
	mux.HandleFunc("GET /plans", h.handleGetPlans)
	mux.HandleFunc("GET /plans/{id}", h.handleGetPlan)
	mux.HandleFunc("POST /plans/{id}/expenses", h.handleAddExpense)
}

// GET /plans
func (h *Handler) handleGetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.store.GetAll()
	if err != nil {
		http.Error(w, "unable to get all plans", http.StatusInternalServerError)
		return
	}

	response := make([]PlanResponse, 0)
	for _, plan := range plans {
		response = append(response, mapToDTO(plan))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /plans
func (h *Handler) handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	id := h.store.GenerateID()
	plan, err := domain.NewPlan(id, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.Save(plan)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapToDTO(plan))
}

// GET /plans/{id}
func (h *Handler) handleGetPlan(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID format", http.StatusBadRequest)
		return
	}

	plan, err := h.store.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToDTO(plan))
}

// POST /plans/{id}/expenses
func (h *Handler) handleAddExpense(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID format", http.StatusBadRequest)
		return
	}

	var req AddExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	plan, err := h.store.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Update the domain object
	if err := plan.AddExpense(req.Name, domain.Money(req.Amount)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save back to store
	h.store.Save(plan)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToDTO(plan))
}
