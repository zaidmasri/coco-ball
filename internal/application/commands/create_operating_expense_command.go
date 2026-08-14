package commands

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateOperatingExpense is the command that finalizes an Operating
// Expenses wizard item into a valid domain.Cost for the first time.
// OperatingExpensesService.CreateOperatingExpense turns this into a
// domain.Cost via domain.NewCost (validating the accumulated wizard field
// values) and assigns it ItemID — the wizard row's pre-existing ID from
// CreateOperatingExpenseDraft — via Cost.SetID, so the domain entity and its
// wizard row share one identity. Callers must not construct the entity
// themselves.
type CreateOperatingExpense struct {
	ItemID             uuid.UUID
	Name               string
	BaseAmountPerMonth domain.Money
	Growth             domain.GrowthStrategy
	CurrentStep        int
}

// CreateOperatingExpenseResult wraps the newly-completed item's Result.
type CreateOperatingExpenseResult struct {
	Result *common.OperatingExpenseResult
}
