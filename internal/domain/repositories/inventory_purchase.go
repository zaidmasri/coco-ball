package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type InventoryPurchaseRepository interface {
	CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindInventoryPurchaseDraft(planID uuid.UUID) (*InventoryPurchaseItem, error)
	GetInventoryPurchase(itemID uuid.UUID) (*InventoryPurchaseItem, error)
	SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status ItemStatus) error
	ListCompleteInventoryPurchases(planID uuid.UUID) ([]InventoryPurchaseItem, error)
	DeleteInventoryPurchase(itemID uuid.UUID) error
}
