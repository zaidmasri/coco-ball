package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type BenefitRepository interface {
	CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error)
	FindBenefitDraft(planID uuid.UUID) (*BenefitItem, error)
	GetBenefit(itemID uuid.UUID) (*BenefitItem, error)
	// SaveBenefitDraftStep persists an in-progress, possibly incomplete
	// wizard step. Drafts are allowed to be invalid by design, so this
	// accepts a bare domain.Benefit.
	SaveBenefitDraftStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int) error
	// CompleteBenefit marks a benefit item complete. It requires a
	// domain.ValidatedBenefit, so it is a compile error to complete an item
	// without validating it first - see domain.NewValidatedBenefit.
	CompleteBenefit(itemID uuid.UUID, benefit domain.ValidatedBenefit, currentStep int) error
	ListCompleteBenefits(planID uuid.UUID) ([]BenefitItem, error)
	DeleteBenefit(itemID uuid.UUID) error
}
