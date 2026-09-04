package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// DistributionResult is the write-side acknowledgement shape for
// CreateDistribution/UpdateDistribution. Mirrors SalaryRoleResult.
type DistributionResult struct {
	ID            uuid.UUID
	Name          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
}
