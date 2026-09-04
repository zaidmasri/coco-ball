package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateStartupCost is the command to change an already-complete Startup
// Costs item's fields. StartingPointService.UpdateStartupCost re-validates
// the fields via domain.NewStartupCost and assigns the result ItemID —
// callers must not construct the entity themselves.
type UpdateStartupCost struct {
	ItemID      uuid.UUID
	Name        string
	Amount      domain.Money
	CurrentStep int
}

// UpdateStartupCostResult wraps the updated item's Result.
type UpdateStartupCostResult struct {
	Result *common.StartupCostResult
}
