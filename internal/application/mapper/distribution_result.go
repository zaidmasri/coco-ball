package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewDistributionResultFromEntity builds a DistributionResult from a
// Distribution value object.
func NewDistributionResultFromEntity(d domain.Distribution) *common.DistributionResult {
	return &common.DistributionResult{
		ID:            d.ID(),
		Name:          d.Name,
		MonthlyAmount: d.MonthlyAmount,
		Growth:        d.GrowthAfterYr1,
	}
}
