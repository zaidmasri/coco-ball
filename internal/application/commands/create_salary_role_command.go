package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateSalaryRole is the command that finalizes a Salary Roles wizard item
// into a valid domain.SalaryRole for the first time. PayrollService.
// CreateSalaryRole turns this into a domain.SalaryRole via
// domain.NewSalaryRole (validating the accumulated wizard field values) and
// assigns it ItemID — the wizard row's pre-existing ID from
// CreateSalaryRoleDraft — via SalaryRole.SetID, so the domain entity and its
// wizard row share one identity. Callers must not construct the entity
// themselves.
type CreateSalaryRole struct {
	ItemID       uuid.UUID
	Role         string
	IsContractor bool
	Headcount    int
	MonthlyPay   domain.Money
	Growth       domain.AnnualGrowth
	CurrentStep  int
}

// CreateSalaryRoleResult wraps the newly-completed item's Result.
type CreateSalaryRoleResult struct {
	Result *common.SalaryRoleResult
}
