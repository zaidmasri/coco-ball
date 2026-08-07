package interfaces

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// CashFlowSummaryResult aggregates what the hub summary page needs.
type CashFlowSummaryResult struct {
	SectionStatus         map[string]bool
	InventoryPurchaseCount int
	DistributionCount     int
}

type CashFlowService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (CashFlowSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Inventory Purchases
	CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error)
	GetInventoryPurchaseDraft(planID uuid.UUID) (*repositories.InventoryPurchaseItem, error)
	GetInventoryPurchase(itemID uuid.UUID) (*repositories.InventoryPurchaseItem, error)
	SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status repositories.ItemStatus) error
	ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error)
	DeleteInventoryPurchase(itemID uuid.UUID) error

	// Distributions
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	GetDistributionDraft(planID uuid.UUID) (*repositories.DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*repositories.DistributionItem, error)
	SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status repositories.ItemStatus) error
	ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error
}
