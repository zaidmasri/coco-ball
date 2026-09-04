package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewStartupCostResultFromEntity builds a StartupCostResult from a
// StartupCost value object.
func NewStartupCostResultFromEntity(s domain.StartupCost) *common.StartupCostResult {
	return &common.StartupCostResult{
		ID:     s.ID(),
		Name:   s.Name,
		Amount: s.Amount,
	}
}
