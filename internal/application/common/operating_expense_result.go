package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// OperatingExpenseResult is the write-side acknowledgement shape for
// CreateOperatingExpense/UpdateOperatingExpense. It deliberately does not
// carry the repository-layer OperatingExpenseItem (Status, CurrentStep,
// mutation methods) — read paths that need the full wizard row (the list
// page, the step pages) go through OperatingExpensesService.
// GetOperatingExpense/ListCompleteOperatingExpenses instead, which return
// repositories.OperatingExpenseItem directly. See AGENTS.md's "Application
// Layer: go-ddd Comparison" section for why PlanResult follows the same
// split.
type OperatingExpenseResult struct {
	ID                 uuid.UUID
	Name               string
	BaseAmountPerMonth domain.Money
	Growth             domain.GrowthStrategy
}
