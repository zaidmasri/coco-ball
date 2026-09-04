package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// StartupCostResult is the write-side acknowledgement shape for
// CreateStartupCost/UpdateStartupCost. Mirrors SalaryRoleResult.
type StartupCostResult struct {
	ID     uuid.UUID
	Name   string
	Amount domain.Money
}
