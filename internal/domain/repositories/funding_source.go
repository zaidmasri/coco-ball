package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type FundingSourceRepository interface {
	CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error)
	FindFundingSourceDraft(planID uuid.UUID) (*FundingSourceItem, error)
	GetFundingSource(itemID uuid.UUID) (*FundingSourceItem, error)
	// SaveFundingSourceDraftStep persists an in-progress, possibly
	// incomplete wizard step. Drafts are allowed to be invalid by design,
	// so this accepts a bare domain.FundingSource.
	SaveFundingSourceDraftStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int) error
	// CompleteFundingSource marks a funding source item complete. It
	// requires a domain.ValidatedFundingSource, so it is a compile error
	// to complete an item without validating it first - see
	// domain.NewValidatedFundingSource.
	CompleteFundingSource(itemID uuid.UUID, funding domain.ValidatedFundingSource, currentStep int) error
	ListCompleteFundingSources(planID uuid.UUID) ([]FundingSourceItem, error)
	DeleteFundingSource(itemID uuid.UUID) error
}
