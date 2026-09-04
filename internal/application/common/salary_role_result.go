package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// SalaryRoleResult is the write-side acknowledgement shape for
// CreateSalaryRole/UpdateSalaryRole. It deliberately does not carry the
// repository-layer SalaryRoleItem (Status, CurrentStep) — read paths that
// need the full wizard row go through PayrollService.GetSalaryRole/
// ListCompleteSalaryRoles instead. Mirrors OperatingExpenseResult.
type SalaryRoleResult struct {
	ID           uuid.UUID
	Role         string
	IsContractor bool
	Headcount    int
	MonthlyPay   domain.Money
	Growth       domain.AnnualGrowth
}
