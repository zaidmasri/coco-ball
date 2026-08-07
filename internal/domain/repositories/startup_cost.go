package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type StartupCostRepository interface {
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	GetStartupCostDraft(planID uuid.UUID) (*StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*StartupCostItem, error)
	SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status ItemStatus) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error
}
