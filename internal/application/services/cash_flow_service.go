package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
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
func (s *cashFlowService) SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status repositories.ItemStatus) error {
	return s.inventoryPurchases.SaveInventoryPurchaseStep(itemID, inv, currentStep, status)
}
func (s *cashFlowService) ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error) {
	return s.inventoryPurchases.ListCompleteInventoryPurchases(planID)
}
func (s *cashFlowService) DeleteInventoryPurchase(itemID uuid.UUID) error {
	return s.inventoryPurchases.DeleteInventoryPurchase(itemID)
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
func (s *cashFlowService) SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status repositories.ItemStatus) error {
	return s.distributions.SaveDistributionStep(itemID, dist, currentStep, status)
}
func (s *cashFlowService) ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error) {
	return s.distributions.ListCompleteDistributions(planID)
}
func (s *cashFlowService) DeleteDistribution(itemID uuid.UUID) error {
	return s.distributions.DeleteDistribution(itemID)
}
