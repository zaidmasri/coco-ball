package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateStartupCost is the command that finalizes a Startup Costs wizard
// item into a valid domain.StartupCost for the first time.
// StartingPointService.CreateStartupCost turns this into a
// domain.StartupCost via domain.NewStartupCost (validating the
// accumulated wizard field values) and assigns it ItemID — the wizard
// row's pre-existing ID from CreateStartupCostDraft — via
// StartupCost.SetID, so the domain entity and its wizard row share one
// identity. Callers must not construct the entity themselves.
type CreateStartupCost struct {
	ItemID      uuid.UUID
	Name        string
	Amount      domain.Money
	CurrentStep int
}

// CreateStartupCostResult wraps the newly-completed item's Result.
type CreateStartupCostResult struct {
	Result *common.StartupCostResult
}
