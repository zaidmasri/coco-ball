package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateFundingSource is the command that finalizes a Funding Sources
// wizard item into a valid domain.FundingSource for the first time.
// StartingPointService.CreateFundingSource turns this into a
// domain.FundingSource via domain.NewFundingSource (validating the
// accumulated wizard field values) and assigns it ItemID — the wizard
// row's pre-existing ID from CreateFundingSourceDraft — via
// FundingSource.SetID, so the domain entity and its wizard row share one
// identity. Callers must not construct the entity themselves.
type CreateFundingSource struct {
	ItemID       uuid.UUID
	Name         string
	Amount       domain.Money
	InterestRate float64
	TermMonths   int
	CurrentStep  int
}

// CreateFundingSourceResult wraps the newly-completed item's Result.
type CreateFundingSourceResult struct {
	Result *common.FundingSourceResult
}
