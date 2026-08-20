package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type DistributionRepository interface {
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	FindDistributionDraft(planID uuid.UUID) (*DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*DistributionItem, error)
	SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status ItemStatus) error
	ListCompleteDistributions(planID uuid.UUID) ([]DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error
}
