package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// BenefitResult is the write-side acknowledgement shape for
// CreateBenefit/UpdateBenefit. Mirrors OperatingExpenseResult.
type BenefitResult struct {
	ID            uuid.UUID
	Type          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
}
