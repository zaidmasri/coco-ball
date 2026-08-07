package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type FundingSourceRepository interface {
	CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error)
	GetFundingSourceDraft(planID uuid.UUID) (*FundingSourceItem, error)
	GetFundingSource(itemID uuid.UUID) (*FundingSourceItem, error)
	SaveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status ItemStatus) error
	ListCompleteFundingSources(planID uuid.UUID) ([]FundingSourceItem, error)
	DeleteFundingSource(itemID uuid.UUID) error
}
