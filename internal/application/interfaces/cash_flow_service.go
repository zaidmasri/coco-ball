package interfaces

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// CashFlowSummaryResult aggregates what the hub summary page needs.
type CashFlowSummaryResult struct {
	SectionStatus          map[string]bool
	InventoryPurchaseCount int
	DistributionCount      int
}

type CashFlowService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (CashFlowSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Inventory Purchases
	CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindInventoryPurchaseDraft(planID uuid.UUID) (*repositories.InventoryPurchaseItem, error)
	GetInventoryPurchase(itemID uuid.UUID) (*repositories.InventoryPurchaseItem, error)
	SaveInventoryPurchaseDraftStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int) error
	ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error)
	DeleteInventoryPurchase(itemID uuid.UUID) error

	// CreateInventoryPurchase/UpdateInventoryPurchase are the only code
	// allowed to construct or mutate the domain.InventoryPurchase entity -
	// see plan_cash_flow.go's NewInventoryPurchase.
	CreateInventoryPurchase(cmd *commands.CreateInventoryPurchase) (*commands.CreateInventoryPurchaseResult, error)
	UpdateInventoryPurchase(cmd *commands.UpdateInventoryPurchase) (*commands.UpdateInventoryPurchaseResult, error)

	// Distributions
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	FindDistributionDraft(planID uuid.UUID) (*repositories.DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*repositories.DistributionItem, error)
	SaveDistributionDraftStep(itemID uuid.UUID, dist domain.Distribution, currentStep int) error
	ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error

	// CreateDistribution/UpdateDistribution are the only code allowed to
	// construct or mutate the domain.Distribution entity - see
	// plan_cash_flow.go's NewDistribution.
	CreateDistribution(cmd *commands.CreateDistribution) (*commands.CreateDistributionResult, error)
	UpdateDistribution(cmd *commands.UpdateDistribution) (*commands.UpdateDistributionResult, error)
}
