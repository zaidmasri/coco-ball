package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateDistribution is the command to change an already-complete
// Distributions item's fields. CashFlowService.UpdateDistribution
// re-validates the fields via domain.NewDistribution and assigns the
// result ItemID — callers must not construct the entity themselves.
type UpdateDistribution struct {
	ItemID        uuid.UUID
	Name          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// UpdateDistributionResult wraps the updated item's Result.
type UpdateDistributionResult struct {
	Result *common.DistributionResult
}
