// Package api handles external requests for the system
package api

import "github.com/zaidmasri/business-planning-tool/internal/domain"

type CreatePlanRequest struct {
	Name           string `json:"name"`
	DurationMonths int    `json:"duration_months"`
}

type FinancialEntryRequest struct {
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
	MonthIndex int    `json:"duration_months"`
}

type FinancialEntryResponse struct {
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
	MonthIndex int    `json:"month_index"`
}

type PlanResponse struct {
	ID             int                      `json:"id"`
	Name           string                   `json:"name"`
	DurationMonths int                      `json:"duration_months"`
	Expenses       []FinancialEntryResponse `json:"expenses"`
	Revenues       []FinancialEntryResponse `json:"revenues"`
}

func mapToDTO(p *domain.Plan) PlanResponse {
	resp := PlanResponse{
		ID:             p.ID(),
		Name:           p.Name(),
		DurationMonths: p.Duration(),
		Expenses:       make([]FinancialEntryResponse, 0),
		Revenues:       make([]FinancialEntryResponse, 0),
	}

	for _, exp := range p.Expenses() {
		resp.Expenses = append(resp.Expenses, FinancialEntryResponse{
			Name:       exp.Name,
			Amount:     int64(exp.Amount),
			MonthIndex: int(exp.Month),
		})
	}

	for _, rev := range p.Revenues() {
		resp.Revenues = append(resp.Revenues, FinancialEntryResponse{
			Name:       rev.Name,
			Amount:     int64(rev.Amount),
			MonthIndex: int(rev.Month),
		})
	}
	return resp
}
