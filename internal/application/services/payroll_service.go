package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type payrollService struct {
	salaryRoles     repositories.SalaryRoleRepository
	benefits        repositories.BenefitRepository
	payrollTaxRates repositories.PayrollTaxRatesRepository
	wizardProgress  repositories.WizardProgressRepository
}

func NewPayrollService(
	salaryRoles repositories.SalaryRoleRepository,
	benefits repositories.BenefitRepository,
	payrollTaxRates repositories.PayrollTaxRatesRepository,
	wizardProgress repositories.WizardProgressRepository,
) interfaces.PayrollService {
	return &payrollService{
		salaryRoles:     salaryRoles,
		benefits:        benefits,
		payrollTaxRates: payrollTaxRates,
		wizardProgress:  wizardProgress,
	}
}

func (s *payrollService) GetHubStatus(planID uuid.UUID) (map[string]bool, error) {
	return s.wizardProgress.GetWizardSectionStatus(planID, domain.HubPayroll)
}

func (s *payrollService) MarkWizardSectionComplete(planID uuid.UUID, section string) error {
	return s.wizardProgress.MarkWizardSectionComplete(planID, domain.HubPayroll, section)
}

func (s *payrollService) GetSummary(planID uuid.UUID) (interfaces.PayrollSummaryResult, error) {
	sectionStatus, err := s.wizardProgress.GetWizardSectionStatus(planID, domain.HubPayroll)
	if err != nil {
		sectionStatus = map[string]bool{}
	}
	roles, err := s.salaryRoles.ListCompleteSalaryRoles(planID)
	if err != nil {
		return interfaces.PayrollSummaryResult{}, err
	}
	bens, err := s.benefits.ListCompleteBenefits(planID)
	if err != nil {
		return interfaces.PayrollSummaryResult{}, err
	}
	return interfaces.PayrollSummaryResult{
		SectionStatus:   sectionStatus,
		SalaryRoleCount: len(roles),
		BenefitCount:    len(bens),
	}, nil
}

func (s *payrollService) CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.salaryRoles.CreateSalaryRoleDraft(planID)
}
func (s *payrollService) FindSalaryRoleDraft(planID uuid.UUID) (*repositories.SalaryRoleItem, error) {
	return s.salaryRoles.FindSalaryRoleDraft(planID)
}
func (s *payrollService) GetSalaryRole(itemID uuid.UUID) (*repositories.SalaryRoleItem, error) {
	return s.salaryRoles.GetSalaryRole(itemID)
}
func (s *payrollService) SaveSalaryRoleDraftStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int) error {
	return s.salaryRoles.SaveSalaryRoleDraftStep(itemID, role, currentStep)
}
func (s *payrollService) ListCompleteSalaryRoles(planID uuid.UUID) ([]repositories.SalaryRoleItem, error) {
	return s.salaryRoles.ListCompleteSalaryRoles(planID)
}
func (s *payrollService) DeleteSalaryRole(itemID uuid.UUID) error {
	return s.salaryRoles.DeleteSalaryRole(itemID)
}

func (s *payrollService) CreateSalaryRole(cmd *commands.CreateSalaryRole) (*commands.CreateSalaryRoleResult, error) {
	role, err := domain.NewSalaryRole(cmd.Role, cmd.IsContractor, cmd.Headcount, cmd.MonthlyPay, cmd.Growth)
	if err != nil {
		return nil, err
	}
	role.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedSalaryRole(role)
	if err != nil {
		return nil, err
	}

	if err := s.salaryRoles.CompleteSalaryRole(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateSalaryRoleResult{Result: mapper.NewSalaryRoleResultFromEntity(role)}, nil
}

func (s *payrollService) UpdateSalaryRole(cmd *commands.UpdateSalaryRole) (*commands.UpdateSalaryRoleResult, error) {
	role, err := domain.NewSalaryRole(cmd.Role, cmd.IsContractor, cmd.Headcount, cmd.MonthlyPay, cmd.Growth)
	if err != nil {
		return nil, err
	}
	role.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedSalaryRole(role)
	if err != nil {
		return nil, err
	}

	if err := s.salaryRoles.CompleteSalaryRole(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateSalaryRoleResult{Result: mapper.NewSalaryRoleResultFromEntity(role)}, nil
}

func (s *payrollService) CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.benefits.CreateBenefitDraft(planID)
}
func (s *payrollService) FindBenefitDraft(planID uuid.UUID) (*repositories.BenefitItem, error) {
	return s.benefits.FindBenefitDraft(planID)
}
func (s *payrollService) GetBenefit(itemID uuid.UUID) (*repositories.BenefitItem, error) {
	return s.benefits.GetBenefit(itemID)
}
func (s *payrollService) SaveBenefitDraftStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int) error {
	return s.benefits.SaveBenefitDraftStep(itemID, benefit, currentStep)
}
func (s *payrollService) ListCompleteBenefits(planID uuid.UUID) ([]repositories.BenefitItem, error) {
	return s.benefits.ListCompleteBenefits(planID)
}
func (s *payrollService) DeleteBenefit(itemID uuid.UUID) error {
	return s.benefits.DeleteBenefit(itemID)
}

func (s *payrollService) CreateBenefit(cmd *commands.CreateBenefit) (*commands.CreateBenefitResult, error) {
	benefit, err := domain.NewBenefit(cmd.Type, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	benefit.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedBenefit(benefit)
	if err != nil {
		return nil, err
	}

	if err := s.benefits.CompleteBenefit(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateBenefitResult{Result: mapper.NewBenefitResultFromEntity(benefit)}, nil
}

func (s *payrollService) UpdateBenefit(cmd *commands.UpdateBenefit) (*commands.UpdateBenefitResult, error) {
	benefit, err := domain.NewBenefit(cmd.Type, cmd.MonthlyAmount, cmd.Growth)
	if err != nil {
		return nil, err
	}
	benefit.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedBenefit(benefit)
	if err != nil {
		return nil, err
	}

	if err := s.benefits.CompleteBenefit(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateBenefitResult{Result: mapper.NewBenefitResultFromEntity(benefit)}, nil
}

func (s *payrollService) GetPayrollTaxRatesRow(planID uuid.UUID) (*repositories.PayrollTaxRatesRow, error) {
	return s.payrollTaxRates.GetPayrollTaxRatesRow(planID)
}
func (s *payrollService) SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int, isComplete bool) error {
	if isComplete {
		if err := domain.ValidatePayrollTaxRates(rates); err != nil {
			return err
		}
	}
	return s.payrollTaxRates.SavePayrollTaxRatesStep(planID, rates, currentStep)
}
