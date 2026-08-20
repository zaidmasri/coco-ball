package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
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
func (s *operatingExpensesService) FindOperatingExpenseDraft(planID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	return s.operatingExpenses.FindOperatingExpenseDraft(planID)
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

func (s *operatingExpensesService) CreateOperatingExpense(cmd *commands.CreateOperatingExpense) (*commands.CreateOperatingExpenseResult, error) {
	cost, err := domain.NewCost(cmd.Name, cmd.BaseAmountPerMonth, cmd.Growth)
	if err != nil {
		return nil, err
	}
	cost.SetID(cmd.ItemID)

	if err := s.operatingExpenses.SaveOperatingExpenseStep(cmd.ItemID, cost, cmd.CurrentStep, repositories.StatusComplete); err != nil {
		return nil, err
	}

	return &commands.CreateOperatingExpenseResult{Result: mapper.NewOperatingExpenseResultFromEntity(cost)}, nil
}

func (s *operatingExpensesService) UpdateOperatingExpense(cmd *commands.UpdateOperatingExpense) (*commands.UpdateOperatingExpenseResult, error) {
	cost, err := domain.NewCost(cmd.Name, cmd.BaseAmountPerMonth, cmd.Growth)
	if err != nil {
		return nil, err
	}
	cost.SetID(cmd.ItemID)

	if err := s.operatingExpenses.SaveOperatingExpenseStep(cmd.ItemID, cost, cmd.CurrentStep, repositories.StatusComplete); err != nil {
		return nil, err
	}

	return &commands.UpdateOperatingExpenseResult{Result: mapper.NewOperatingExpenseResultFromEntity(cost)}, nil
}

func (s *operatingExpensesService) DeleteOperatingExpense(cmd *commands.DeleteOperatingExpense) (*commands.DeleteOperatingExpenseResult, error) {
	if err := s.operatingExpenses.DeleteOperatingExpense(cmd.ItemID); err != nil {
		return nil, err
	}
	return &commands.DeleteOperatingExpenseResult{Success: true}, nil
}
