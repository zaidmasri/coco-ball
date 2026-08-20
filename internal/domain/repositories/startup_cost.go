package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type StartupCostRepository interface {
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	FindStartupCostDraft(planID uuid.UUID) (*StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*StartupCostItem, error)
	SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status ItemStatus) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error
}
