package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type operatingExpensesService struct {
	operatingExpenses repositories.OperatingExpenseRepository
	wizardProgress    repositories.WizardProgressRepository
}

func NewOperatingExpensesService(
	operatingExpenses repositories.OperatingExpenseRepository,
	wizardProgress repositories.WizardProgressRepository,
) interfaces.OperatingExpensesService {
	return &operatingExpensesService{
		operatingExpenses: operatingExpenses,
		wizardProgress:    wizardProgress,
	}
}

func (s *operatingExpensesService) GetHubStatus(planID uuid.UUID) (map[string]bool, error) {
	return s.wizardProgress.GetWizardSectionStatus(planID, domain.HubOperatingExpenses)
}

func (s *operatingExpensesService) MarkWizardSectionComplete(planID uuid.UUID, section string) error {
	return s.wizardProgress.MarkWizardSectionComplete(planID, domain.HubOperatingExpenses, section)
}

func (s *operatingExpensesService) GetSummary(planID uuid.UUID) (interfaces.OperatingExpensesSummaryResult, error) {
	sectionStatus, err := s.wizardProgress.GetWizardSectionStatus(planID, domain.HubOperatingExpenses)
	if err != nil {
		sectionStatus = map[string]bool{}
	}
	expenses, err := s.operatingExpenses.ListCompleteOperatingExpenses(planID)
	if err != nil {
		return interfaces.OperatingExpensesSummaryResult{}, err
	}
	return interfaces.OperatingExpensesSummaryResult{
		SectionStatus:         sectionStatus,
		OperatingExpenseCount: len(expenses),
	}, nil
}

func (s *operatingExpensesService) CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.operatingExpenses.CreateOperatingExpenseDraft(planID)
}
func (s *operatingExpensesService) GetOperatingExpenseDraft(planID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	return s.operatingExpenses.GetOperatingExpenseDraft(planID)
}
func (s *operatingExpensesService) GetOperatingExpense(itemID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	return s.operatingExpenses.GetOperatingExpense(itemID)
}
func (s *operatingExpensesService) SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status repositories.ItemStatus) error {
	return s.operatingExpenses.SaveOperatingExpenseStep(itemID, cost, currentStep, status)
}
func (s *operatingExpensesService) ListCompleteOperatingExpenses(planID uuid.UUID) ([]repositories.OperatingExpenseItem, error) {
	return s.operatingExpenses.ListCompleteOperatingExpenses(planID)
}
func (s *operatingExpensesService) DeleteOperatingExpense(itemID uuid.UUID) error {
	return s.operatingExpenses.DeleteOperatingExpense(itemID)
}
