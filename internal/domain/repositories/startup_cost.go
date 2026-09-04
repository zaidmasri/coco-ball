package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type StartupCostRepository interface {
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	FindStartupCostDraft(planID uuid.UUID) (*StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*StartupCostItem, error)
	// SaveStartupCostDraftStep persists an in-progress, possibly
	// incomplete wizard step. Drafts are allowed to be invalid by design,
	// so this accepts a bare domain.StartupCost.
	SaveStartupCostDraftStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int) error
	// CompleteStartupCost marks a startup cost item complete. It requires
	// a domain.ValidatedStartupCost, so it is a compile error to complete
	// an item without validating it first - see domain.NewValidatedStartupCost.
	CompleteStartupCost(itemID uuid.UUID, cost domain.ValidatedStartupCost, currentStep int) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error
}
