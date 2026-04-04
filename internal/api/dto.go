// Package api handles external requests for the system
package api

import "github.com/zaidmasri/business-planning-tool/internal/domain"

type CreatePlanRequest struct {
	Name string `json:"name"`
}

type AddExpenseRequest struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

type ExpenseResponse struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

type PlanResponse struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	TotalExpenses int64             `json:"total_expenses"`
	Expenses      []ExpenseResponse `json:"expenses"`
}

// mapToDTO is a helper to convert our encapsulated domain model into public JSON.
func mapToDTO(p *domain.Plan) PlanResponse {
	resp := PlanResponse{
		ID:            p.ID(),
		Name:          p.Name(),
		TotalExpenses: int64(p.TotalExpenses()),
		Expenses:      make([]ExpenseResponse, 0),
	}

	for _, exp := range p.Expenses() {
		resp.Expenses = append(resp.Expenses, ExpenseResponse{
			Name:   exp.Name,
			Amount: int64(exp.Amount),
		})
	}
	return resp
}
