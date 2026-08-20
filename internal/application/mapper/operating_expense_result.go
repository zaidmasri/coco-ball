package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewOperatingExpenseResultFromEntity builds an OperatingExpenseResult from
// a Cost value object.
func NewOperatingExpenseResultFromEntity(c domain.Cost) *common.OperatingExpenseResult {
	return &common.OperatingExpenseResult{
		ID:                 c.ID(),
		Name:               c.Name(),
		BaseAmountPerMonth: c.BaseAmountPerMonth(),
		Growth:             c.Growth(),
	}
}
