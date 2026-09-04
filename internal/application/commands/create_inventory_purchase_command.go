package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateInventoryPurchase is the command that finalizes an Inventory
// Purchases wizard item into a valid domain.InventoryPurchase for the first
// time. CashFlowService.CreateInventoryPurchase turns this into a
// domain.InventoryPurchase via domain.NewInventoryPurchase (validating the
// accumulated wizard field values) and assigns it ItemID — the wizard row's
// pre-existing ID from CreateInventoryPurchaseDraft — via
// InventoryPurchase.SetID, so the domain entity and its wizard row share
// one identity. Callers must not construct the entity themselves.
type CreateInventoryPurchase struct {
	ItemID        uuid.UUID
	Category      string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// CreateInventoryPurchaseResult wraps the newly-completed item's Result.
type CreateInventoryPurchaseResult struct {
	Result *common.InventoryPurchaseResult
}
