package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type cashFlowService struct {
	inventoryPurchases repositories.InventoryPurchaseRepository
	distributions      repositories.DistributionRepository
	wizardProgress     repositories.WizardProgressRepository
}

func NewCashFlowService(
	inventoryPurchases repositories.InventoryPurchaseRepository,
	distributions repositories.DistributionRepository,
	wizardProgress repositories.WizardProgressRepository,
) interfaces.CashFlowService {
	return &cashFlowService{
		inventoryPurchases: inventoryPurchases,
		distributions:      distributions,
		wizardProgress:     wizardProgress,
	}
}

func (s *cashFlowService) GetHubStatus(planID uuid.UUID) (map[string]bool, error) {
	return s.wizardProgress.GetWizardSectionStatus(planID, domain.HubCashFlow)
}

func (s *cashFlowService) MarkWizardSectionComplete(planID uuid.UUID, section string) error {
	return s.wizardProgress.MarkWizardSectionComplete(planID, domain.HubCashFlow, section)
}

func (s *cashFlowService) GetSummary(planID uuid.UUID) (interfaces.CashFlowSummaryResult, error) {
	sectionStatus, err := s.wizardProgress.GetWizardSectionStatus(planID, domain.HubCashFlow)
	if err != nil {
		sectionStatus = map[string]bool{}
	}
	invItems, err := s.inventoryPurchases.ListCompleteInventoryPurchases(planID)
	if err != nil {
		return interfaces.CashFlowSummaryResult{}, err
	}
	distItems, err := s.distributions.ListCompleteDistributions(planID)
	if err != nil {
		return interfaces.CashFlowSummaryResult{}, err
	}
	return interfaces.CashFlowSummaryResult{
		SectionStatus:          sectionStatus,
		InventoryPurchaseCount: len(invItems),
		DistributionCount:      len(distItems),
	}, nil
}

func (s *cashFlowService) CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.inventoryPurchases.CreateInventoryPurchaseDraft(planID)
}
func (s *cashFlowService) FindInventoryPurchaseDraft(planID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	return s.inventoryPurchases.FindInventoryPurchaseDraft(planID)
}
func (s *cashFlowService) GetInventoryPurchase(itemID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	return s.inventoryPurchases.GetInventoryPurchase(itemID)
}
func (s *cashFlowService) SaveInventoryPurchaseDraftStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int) error {
	return s.inventoryPurchases.SaveInventoryPurchaseDraftStep(itemID, inv, currentStep)
}
func (s *cashFlowService) ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error) {
	return s.inventoryPurchases.ListCompleteInventoryPurchases(planID)
}
func (s *cashFlowService) DeleteInventoryPurchase(itemID uuid.UUID) error {
	return s.inventoryPurchases.DeleteInventoryPurchase(itemID)
}

func (s *cashFlowService) CreateInventoryPurchase(cmd *commands.CreateInventoryPurchase) (*commands.CreateInventoryPurchaseResult, error) {
	inv, err := domain.NewInventoryPurchase(cmd.Category, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	inv.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedInventoryPurchase(inv)
	if err != nil {
		return nil, err
	}

	if err := s.inventoryPurchases.CompleteInventoryPurchase(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateInventoryPurchaseResult{Result: mapper.NewInventoryPurchaseResultFromEntity(inv)}, nil
}

func (s *cashFlowService) UpdateInventoryPurchase(cmd *commands.UpdateInventoryPurchase) (*commands.UpdateInventoryPurchaseResult, error) {
	inv, err := domain.NewInventoryPurchase(cmd.Category, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	inv.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedInventoryPurchase(inv)
	if err != nil {
		return nil, err
	}

	if err := s.inventoryPurchases.CompleteInventoryPurchase(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateInventoryPurchaseResult{Result: mapper.NewInventoryPurchaseResultFromEntity(inv)}, nil
}

func (s *cashFlowService) CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.distributions.CreateDistributionDraft(planID)
}
func (s *cashFlowService) FindDistributionDraft(planID uuid.UUID) (*repositories.DistributionItem, error) {
	return s.distributions.FindDistributionDraft(planID)
}
func (s *cashFlowService) GetDistribution(itemID uuid.UUID) (*repositories.DistributionItem, error) {
	return s.distributions.GetDistribution(itemID)
}
func (s *cashFlowService) SaveDistributionDraftStep(itemID uuid.UUID, dist domain.Distribution, currentStep int) error {
	return s.distributions.SaveDistributionDraftStep(itemID, dist, currentStep)
}
func (s *cashFlowService) ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error) {
	return s.distributions.ListCompleteDistributions(planID)
}
func (s *cashFlowService) DeleteDistribution(itemID uuid.UUID) error {
	return s.distributions.DeleteDistribution(itemID)
}

func (s *cashFlowService) CreateDistribution(cmd *commands.CreateDistribution) (*commands.CreateDistributionResult, error) {
	dist, err := domain.NewDistribution(cmd.Name, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	dist.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedDistribution(dist)
	if err != nil {
		return nil, err
	}

	if err := s.distributions.CompleteDistribution(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateDistributionResult{Result: mapper.NewDistributionResultFromEntity(dist)}, nil
}

func (s *cashFlowService) UpdateDistribution(cmd *commands.UpdateDistribution) (*commands.UpdateDistributionResult, error) {
	dist, err := domain.NewDistribution(cmd.Name, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	dist.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedDistribution(dist)
	if err != nil {
		return nil, err
	}

	if err := s.distributions.CompleteDistribution(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateDistributionResult{Result: mapper.NewDistributionResultFromEntity(dist)}, nil
}
