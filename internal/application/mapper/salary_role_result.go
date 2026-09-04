package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewSalaryRoleResultFromEntity builds a SalaryRoleResult from a SalaryRole
// value object.
func NewSalaryRoleResultFromEntity(s domain.SalaryRole) *common.SalaryRoleResult {
	return &common.SalaryRoleResult{
		ID:           s.ID(),
		Role:         s.Role,
		IsContractor: s.IsContractor,
		Headcount:    s.Headcount,
		MonthlyPay:   s.MonthlyPay,
		Growth:       s.GrowthAfterYr1,
	}
}
