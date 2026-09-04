package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateSalaryRole is the command to change an already-complete Salary
// Roles item's fields. PayrollService.UpdateSalaryRole re-validates the
// fields via domain.NewSalaryRole and assigns the result ItemID — callers
// must not construct the entity themselves.
type UpdateSalaryRole struct {
	ItemID       uuid.UUID
	Role         string
	IsContractor bool
	Headcount    int
	MonthlyPay   domain.Money
	Growth       domain.AnnualGrowth
	CurrentStep  int
}

// UpdateSalaryRoleResult wraps the updated item's Result.
type UpdateSalaryRoleResult struct {
	Result *common.SalaryRoleResult
}
