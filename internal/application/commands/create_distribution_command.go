package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateDistribution is the command that finalizes a Distributions wizard
// item into a valid domain.Distribution for the first time.
// CashFlowService.CreateDistribution turns this into a domain.Distribution
// via domain.NewDistribution (validating the accumulated wizard field
// values) and assigns it ItemID — the wizard row's pre-existing ID from
// CreateDistributionDraft — via Distribution.SetID, so the domain entity
// and its wizard row share one identity. Callers must not construct the
// entity themselves.
type CreateDistribution struct {
	ItemID        uuid.UUID
	Name          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// CreateDistributionResult wraps the newly-completed item's Result.
type CreateDistributionResult struct {
	Result *common.DistributionResult
}
