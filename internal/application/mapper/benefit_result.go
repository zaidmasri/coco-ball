package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewBenefitResultFromEntity builds a BenefitResult from a Benefit value
// object.
func NewBenefitResultFromEntity(b domain.Benefit) *common.BenefitResult {
	return &common.BenefitResult{
		ID:            b.ID(),
		Type:          b.Type,
		MonthlyAmount: b.MonthlyAmount,
		Growth:        b.GrowthAfterYr1,
	}
}
