package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateOperatingExpense is the command to change an already-complete
// Operating Expenses item's fields. OperatingExpensesService.
// UpdateOperatingExpense re-validates the fields via domain.NewCost and
// assigns the result ItemID — callers must not construct the entity
// themselves.
type UpdateOperatingExpense struct {
	ItemID             uuid.UUID
	Name               string
	BaseAmountPerMonth domain.Money
	Growth             domain.GrowthStrategy
	CurrentStep        int
}

// UpdateOperatingExpenseResult wraps the updated item's Result.
type UpdateOperatingExpenseResult struct {
	Result *common.OperatingExpenseResult
}
