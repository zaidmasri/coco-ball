package interfaces

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
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
	SaveCapitalAssetDraftStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error

	// CreateCapitalAsset/UpdateCapitalAsset are the only code allowed to
	// construct or mutate the domain.CapitalAsset entity - see
	// plan_starting_point.go's NewCapitalAsset.
	CreateCapitalAsset(cmd *commands.CreateCapitalAsset) (*commands.CreateCapitalAssetResult, error)
	UpdateCapitalAsset(cmd *commands.UpdateCapitalAsset) (*commands.UpdateCapitalAssetResult, error)

	// Startup Costs
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	FindStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error)
	SaveStartupCostDraftStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error

	// CreateStartupCost/UpdateStartupCost are the only code allowed to
	// construct or mutate the domain.StartupCost entity - see
	// plan_starting_point.go's NewStartupCost.
	CreateStartupCost(cmd *commands.CreateStartupCost) (*commands.CreateStartupCostResult, error)
	UpdateStartupCost(cmd *commands.UpdateStartupCost) (*commands.UpdateStartupCostResult, error)

	// Funding Sources
	CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error)
	FindFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error)
	GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error)
	SaveFundingSourceDraftStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int) error
	ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error)
	DeleteFundingSource(itemID uuid.UUID) error

	// CreateFundingSource/UpdateFundingSource are the only code allowed to
	// construct or mutate the domain.FundingSource entity - see
	// plan_starting_point.go's NewFundingSource.
	CreateFundingSource(cmd *commands.CreateFundingSource) (*commands.CreateFundingSourceResult, error)
	UpdateFundingSource(cmd *commands.UpdateFundingSource) (*commands.UpdateFundingSourceResult, error)

	// Starting Balances (singleton)
	GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error)
	SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int, isComplete bool) error
}
