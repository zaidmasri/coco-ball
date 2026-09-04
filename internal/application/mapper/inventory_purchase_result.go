package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewInventoryPurchaseResultFromEntity builds an InventoryPurchaseResult
// from an InventoryPurchase value object.
func NewInventoryPurchaseResultFromEntity(i domain.InventoryPurchase) *common.InventoryPurchaseResult {
	return &common.InventoryPurchaseResult{
		ID:            i.ID(),
		Category:      i.Category,
		MonthlyAmount: i.MonthlyAmount,
		Growth:        i.GrowthAfterYr1,
	}
}
