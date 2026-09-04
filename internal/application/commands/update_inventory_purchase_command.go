package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateInventoryPurchase is the command to change an already-complete
// Inventory Purchases item's fields. CashFlowService.UpdateInventoryPurchase
// re-validates the fields via domain.NewInventoryPurchase and assigns the
// result ItemID — callers must not construct the entity themselves.
type UpdateInventoryPurchase struct {
	ItemID        uuid.UUID
	Category      string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// UpdateInventoryPurchaseResult wraps the updated item's Result.
type UpdateInventoryPurchaseResult struct {
	Result *common.InventoryPurchaseResult
}
