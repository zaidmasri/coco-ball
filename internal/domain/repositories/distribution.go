package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type DistributionRepository interface {
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	GetDistributionDraft(planID uuid.UUID) (*DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*DistributionItem, error)
	SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status ItemStatus) error
	ListCompleteDistributions(planID uuid.UUID) ([]DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error
}
