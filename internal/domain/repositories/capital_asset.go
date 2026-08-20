package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type CapitalAssetRepository interface {
	CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error)
	FindCapitalAssetDraft(planID uuid.UUID) (*CapitalAssetItem, error)
	GetCapitalAsset(itemID uuid.UUID) (*CapitalAssetItem, error)
	SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status ItemStatus) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error
}
