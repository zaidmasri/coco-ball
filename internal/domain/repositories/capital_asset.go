package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type CapitalAssetRepository interface {
	CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error)
	GetCapitalAssetDraft(planID uuid.UUID) (*CapitalAssetItem, error)
	GetCapitalAsset(itemID uuid.UUID) (*CapitalAssetItem, error)
	SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status ItemStatus) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error
}
