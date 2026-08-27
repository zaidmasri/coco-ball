package interfaces

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// StartingPointSummaryResult aggregates what the hub summary page needs
// from multiple repository queries into a single service call.
type StartingPointSummaryResult struct {
	SectionStatus      map[string]bool
	FixedAssetCount    int
	StartupCostCount   int
	FundingSourceCount int
}

type StartingPointService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (StartingPointSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Fixed Assets
	CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error)
	FindCapitalAssetDraft(planID uuid.UUID) (*repositories.CapitalAssetItem, error)
	GetCapitalAsset(itemID uuid.UUID) (*repositories.CapitalAssetItem, error)
	SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status repositories.ItemStatus) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error

	// Startup Costs
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	FindStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error)
	SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status repositories.ItemStatus) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error

	// Funding Sources
	CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error)
	FindFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error)
	GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error)
	SaveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status repositories.ItemStatus) error
	ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error)
	DeleteFundingSource(itemID uuid.UUID) error

	// Starting Balances (singleton)
	GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error)
	SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int, isComplete bool) error
}
