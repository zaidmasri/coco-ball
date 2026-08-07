package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
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
func (s *startingPointService) GetCapitalAssetDraft(planID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	return s.capitalAssets.GetCapitalAssetDraft(planID)
}
func (s *startingPointService) GetCapitalAsset(itemID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	return s.capitalAssets.GetCapitalAsset(itemID)
}
func (s *startingPointService) SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status repositories.ItemStatus) error {
	return s.capitalAssets.SaveCapitalAssetStep(itemID, asset, currentStep, status)
}
func (s *startingPointService) ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error) {
	return s.capitalAssets.ListCompleteCapitalAssets(planID)
}
func (s *startingPointService) DeleteCapitalAsset(itemID uuid.UUID) error {
	return s.capitalAssets.DeleteCapitalAsset(itemID)
}

// Startup Costs

func (s *startingPointService) CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.startupCosts.CreateStartupCostDraft(planID)
}
func (s *startingPointService) GetStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error) {
	return s.startupCosts.GetStartupCostDraft(planID)
}
func (s *startingPointService) GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error) {
	return s.startupCosts.GetStartupCost(itemID)
}
func (s *startingPointService) SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status repositories.ItemStatus) error {
	return s.startupCosts.SaveStartupCostStep(itemID, cost, currentStep, status)
}
func (s *startingPointService) ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error) {
	return s.startupCosts.ListCompleteStartupCosts(planID)
}
func (s *startingPointService) DeleteStartupCost(itemID uuid.UUID) error {
	return s.startupCosts.DeleteStartupCost(itemID)
}

// Funding Sources

func (s *startingPointService) CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.fundingSources.CreateFundingSourceDraft(planID)
}
func (s *startingPointService) GetFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error) {
	return s.fundingSources.GetFundingSourceDraft(planID)
}
func (s *startingPointService) GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error) {
	return s.fundingSources.GetFundingSource(itemID)
}
func (s *startingPointService) SaveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status repositories.ItemStatus) error {
	return s.fundingSources.SaveFundingSourceStep(itemID, funding, currentStep, status)
}
func (s *startingPointService) ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error) {
	return s.fundingSources.ListCompleteFundingSources(planID)
}
func (s *startingPointService) DeleteFundingSource(itemID uuid.UUID) error {
	return s.fundingSources.DeleteFundingSource(itemID)
}

// Starting Balances

func (s *startingPointService) GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error) {
	return s.startingBalances.GetStartingBalancesRow(planID)
}
func (s *startingPointService) SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error {
	return s.startingBalances.SaveStartingBalancesStep(planID, balances, currentStep)
}
