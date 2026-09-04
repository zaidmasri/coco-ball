package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateFundingSource is the command to change an already-complete Funding
// Sources item's fields. StartingPointService.UpdateFundingSource
// re-validates the fields via domain.NewFundingSource and assigns the
// result ItemID — callers must not construct the entity themselves.
type UpdateFundingSource struct {
	ItemID       uuid.UUID
	Name         string
	Amount       domain.Money
	InterestRate float64
	TermMonths   int
	CurrentStep  int
}

// UpdateFundingSourceResult wraps the updated item's Result.
type UpdateFundingSourceResult struct {
	Result *common.FundingSourceResult
}
