package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// FundingSourceResult is the write-side acknowledgement shape for
// CreateFundingSource/UpdateFundingSource. Mirrors SalaryRoleResult.
type FundingSourceResult struct {
	ID           uuid.UUID
	Name         string
	Amount       domain.Money
	InterestRate float64
	TermMonths   int
}
