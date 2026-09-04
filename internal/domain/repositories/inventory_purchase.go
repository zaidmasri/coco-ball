package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type InventoryPurchaseRepository interface {
	CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindInventoryPurchaseDraft(planID uuid.UUID) (*InventoryPurchaseItem, error)
	GetInventoryPurchase(itemID uuid.UUID) (*InventoryPurchaseItem, error)
	// SaveInventoryPurchaseDraftStep persists an in-progress, possibly
	// incomplete wizard step. Drafts are allowed to be invalid by design, so
	// this accepts a bare domain.InventoryPurchase.
	SaveInventoryPurchaseDraftStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int) error
	// CompleteInventoryPurchase marks an inventory purchase item complete.
	// It requires a domain.ValidatedInventoryPurchase, so it is a compile
	// error to complete an item without validating it first - see
	// domain.NewValidatedInventoryPurchase.
	CompleteInventoryPurchase(itemID uuid.UUID, inv domain.ValidatedInventoryPurchase, currentStep int) error
	ListCompleteInventoryPurchases(planID uuid.UUID) ([]InventoryPurchaseItem, error)
	DeleteInventoryPurchase(itemID uuid.UUID) error
}
