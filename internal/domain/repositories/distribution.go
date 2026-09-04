package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type DistributionRepository interface {
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	FindDistributionDraft(planID uuid.UUID) (*DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*DistributionItem, error)
	// SaveDistributionDraftStep persists an in-progress, possibly incomplete
	// wizard step. Drafts are allowed to be invalid by design, so this
	// accepts a bare domain.Distribution.
	SaveDistributionDraftStep(itemID uuid.UUID, dist domain.Distribution, currentStep int) error
	// CompleteDistribution marks a distribution item complete. It requires a
	// domain.ValidatedDistribution, so it is a compile error to complete an
	// item without validating it first - see domain.NewValidatedDistribution.
	CompleteDistribution(itemID uuid.UUID, dist domain.ValidatedDistribution, currentStep int) error
	ListCompleteDistributions(planID uuid.UUID) ([]DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error
}
