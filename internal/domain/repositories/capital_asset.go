package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type CapitalAssetRepository interface {
	CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error)
	FindCapitalAssetDraft(planID uuid.UUID) (*CapitalAssetItem, error)
	GetCapitalAsset(itemID uuid.UUID) (*CapitalAssetItem, error)
	// SaveCapitalAssetDraftStep persists an in-progress, possibly
	// incomplete wizard step. Drafts are allowed to be invalid by design,
	// so this accepts a bare domain.CapitalAsset.
	SaveCapitalAssetDraftStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int) error
	// CompleteCapitalAsset marks a capital asset item complete. It
	// requires a domain.ValidatedCapitalAsset, so it is a compile error to
	// complete an item without validating it first - see
	// domain.NewValidatedCapitalAsset.
	CompleteCapitalAsset(itemID uuid.UUID, asset domain.ValidatedCapitalAsset, currentStep int) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error
}
