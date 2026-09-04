package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// InventoryPurchaseResult is the write-side acknowledgement shape for
// CreateInventoryPurchase/UpdateInventoryPurchase. It deliberately does not
// carry the repository-layer InventoryPurchaseItem (Status, CurrentStep) —
// read paths that need the full wizard row go through
// CashFlowService.GetInventoryPurchase/ListCompleteInventoryPurchases
// instead. Mirrors SalaryRoleResult.
type InventoryPurchaseResult struct {
	ID            uuid.UUID
	Category      string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
}
