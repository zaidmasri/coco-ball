package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type startingPointService struct {
	capitalAssets    repositories.CapitalAssetRepository
	startupCosts     repositories.StartupCostRepository
	fundingSources   repositories.FundingSourceRepository
	startingBalances repositories.StartingBalancesRepository
	wizardProgress   repositories.WizardProgressRepository
}

func NewStartingPointService(
	capitalAssets repositories.CapitalAssetRepository,
	startupCosts repositories.StartupCostRepository,
	fundingSources repositories.FundingSourceRepository,
	startingBalances repositories.StartingBalancesRepository,
	wizardProgress repositories.WizardProgressRepository,
) interfaces.StartingPointService {
	return &startingPointService{
		capitalAssets:    capitalAssets,
		startupCosts:     startupCosts,
		fundingSources:   fundingSources,
		startingBalances: startingBalances,
		wizardProgress:   wizardProgress,
	}
}

func (s *startingPointService) GetHubStatus(planID uuid.UUID) (map[string]bool, error) {
	return s.wizardProgress.GetWizardSectionStatus(planID, domain.HubStartingPoint)
}

func (s *startingPointService) MarkWizardSectionComplete(planID uuid.UUID, section string) error {
	return s.wizardProgress.MarkWizardSectionComplete(planID, domain.HubStartingPoint, section)
}

func (s *startingPointService) GetSummary(planID uuid.UUID) (interfaces.StartingPointSummaryResult, error) {
	sectionStatus, err := s.wizardProgress.GetWizardSectionStatus(planID, domain.HubStartingPoint)
	if err != nil {
		sectionStatus = map[string]bool{}
	}
	assets, err := s.capitalAssets.ListCompleteCapitalAssets(planID)
	if err != nil {
		return interfaces.StartingPointSummaryResult{}, err
	}
	costs, err := s.startupCosts.ListCompleteStartupCosts(planID)
	if err != nil {
		return interfaces.StartingPointSummaryResult{}, err
	}
	funding, err := s.fundingSources.ListCompleteFundingSources(planID)
	if err != nil {
		return interfaces.StartingPointSummaryResult{}, err
	}
	return interfaces.StartingPointSummaryResult{
		SectionStatus:      sectionStatus,
		FixedAssetCount:    len(assets),
		StartupCostCount:   len(costs),
		FundingSourceCount: len(funding),
	}, nil
}

// Fixed Assets

func (s *startingPointService) CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.capitalAssets.CreateCapitalAssetDraft(planID)
}
func (s *startingPointService) FindCapitalAssetDraft(planID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	return s.capitalAssets.FindCapitalAssetDraft(planID)
}
func (s *startingPointService) GetCapitalAsset(itemID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	return s.capitalAssets.GetCapitalAsset(itemID)
}
func (s *startingPointService) SaveCapitalAssetDraftStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int) error {
	return s.capitalAssets.SaveCapitalAssetDraftStep(itemID, asset, currentStep)
}
func (s *startingPointService) ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error) {
	return s.capitalAssets.ListCompleteCapitalAssets(planID)
}
func (s *startingPointService) DeleteCapitalAsset(itemID uuid.UUID) error {
	return s.capitalAssets.DeleteCapitalAsset(itemID)
}

func (s *startingPointService) CreateCapitalAsset(cmd *commands.CreateCapitalAsset) (*commands.CreateCapitalAssetResult, error) {
	asset, err := domain.NewCapitalAsset(cmd.Name, cmd.PurchaseCost, cmd.UsefulLifeMonths, cmd.SalvageValue, cmd.PurchaseMonthIndex, cmd.DepreciationMethod, cmd.AssociatedLoan)
	if err != nil {
		return nil, err
	}
	asset.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedCapitalAsset(asset)
	if err != nil {
		return nil, err
	}

	if err := s.capitalAssets.CompleteCapitalAsset(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateCapitalAssetResult{Result: mapper.NewCapitalAssetResultFromEntity(asset)}, nil
}

func (s *startingPointService) UpdateCapitalAsset(cmd *commands.UpdateCapitalAsset) (*commands.UpdateCapitalAssetResult, error) {
	asset, err := domain.NewCapitalAsset(cmd.Name, cmd.PurchaseCost, cmd.UsefulLifeMonths, cmd.SalvageValue, cmd.PurchaseMonthIndex, cmd.DepreciationMethod, cmd.AssociatedLoan)
	if err != nil {
		return nil, err
	}
	asset.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedCapitalAsset(asset)
	if err != nil {
		return nil, err
	}

	if err := s.capitalAssets.CompleteCapitalAsset(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateCapitalAssetResult{Result: mapper.NewCapitalAssetResultFromEntity(asset)}, nil
}

// Startup Costs

func (s *startingPointService) CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.startupCosts.CreateStartupCostDraft(planID)
}
func (s *startingPointService) FindStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error) {
	return s.startupCosts.FindStartupCostDraft(planID)
}
func (s *startingPointService) GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error) {
	return s.startupCosts.GetStartupCost(itemID)
}
func (s *startingPointService) SaveStartupCostDraftStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int) error {
	return s.startupCosts.SaveStartupCostDraftStep(itemID, cost, currentStep)
}
func (s *startingPointService) ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error) {
	return s.startupCosts.ListCompleteStartupCosts(planID)
}
func (s *startingPointService) DeleteStartupCost(itemID uuid.UUID) error {
	return s.startupCosts.DeleteStartupCost(itemID)
}

func (s *startingPointService) CreateStartupCost(cmd *commands.CreateStartupCost) (*commands.CreateStartupCostResult, error) {
	cost, err := domain.NewStartupCost(cmd.Name, cmd.Amount)
	if err != nil {
		return nil, err
	}
	cost.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedStartupCost(cost)
	if err != nil {
		return nil, err
	}

	if err := s.startupCosts.CompleteStartupCost(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateStartupCostResult{Result: mapper.NewStartupCostResultFromEntity(cost)}, nil
}

func (s *startingPointService) UpdateStartupCost(cmd *commands.UpdateStartupCost) (*commands.UpdateStartupCostResult, error) {
	cost, err := domain.NewStartupCost(cmd.Name, cmd.Amount)
	if err != nil {
		return nil, err
	}
	cost.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedStartupCost(cost)
	if err != nil {
		return nil, err
	}

	if err := s.startupCosts.CompleteStartupCost(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateStartupCostResult{Result: mapper.NewStartupCostResultFromEntity(cost)}, nil
}

// Funding Sources

func (s *startingPointService) CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.fundingSources.CreateFundingSourceDraft(planID)
}
func (s *startingPointService) FindFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error) {
	return s.fundingSources.FindFundingSourceDraft(planID)
}
func (s *startingPointService) GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error) {
	return s.fundingSources.GetFundingSource(itemID)
}
func (s *startingPointService) SaveFundingSourceDraftStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int) error {
	return s.fundingSources.SaveFundingSourceDraftStep(itemID, funding, currentStep)
}
func (s *startingPointService) ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error) {
	return s.fundingSources.ListCompleteFundingSources(planID)
}
func (s *startingPointService) DeleteFundingSource(itemID uuid.UUID) error {
	return s.fundingSources.DeleteFundingSource(itemID)
}

func (s *startingPointService) CreateFundingSource(cmd *commands.CreateFundingSource) (*commands.CreateFundingSourceResult, error) {
	funding, err := domain.NewFundingSource(cmd.Name, cmd.Amount, cmd.InterestRate, cmd.TermMonths)
	if err != nil {
		return nil, err
	}
	funding.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedFundingSource(funding)
	if err != nil {
		return nil, err
	}

	if err := s.fundingSources.CompleteFundingSource(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateFundingSourceResult{Result: mapper.NewFundingSourceResultFromEntity(funding)}, nil
}

func (s *startingPointService) UpdateFundingSource(cmd *commands.UpdateFundingSource) (*commands.UpdateFundingSourceResult, error) {
	funding, err := domain.NewFundingSource(cmd.Name, cmd.Amount, cmd.InterestRate, cmd.TermMonths)
	if err != nil {
		return nil, err
	}
	funding.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedFundingSource(funding)
	if err != nil {
		return nil, err
	}

	if err := s.fundingSources.CompleteFundingSource(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateFundingSourceResult{Result: mapper.NewFundingSourceResultFromEntity(funding)}, nil
}

// Starting Balances

func (s *startingPointService) GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error) {
	return s.startingBalances.GetStartingBalancesRow(planID)
}
func (s *startingPointService) SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int, isComplete bool) error {
	if isComplete {
		if err := domain.ValidateStartingBalances(balances); err != nil {
			return err
		}
	}
	return s.startingBalances.SaveStartingBalancesStep(planID, balances, currentStep)
}
