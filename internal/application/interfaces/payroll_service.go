package interfaces

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// PayrollSummaryResult aggregates what the hub summary page needs.
type PayrollSummaryResult struct {
	SectionStatus   map[string]bool
	SalaryRoleCount int
	BenefitCount    int
}

type PayrollService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (PayrollSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Salary Roles
	CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error)
	FindSalaryRoleDraft(planID uuid.UUID) (*repositories.SalaryRoleItem, error)
	GetSalaryRole(itemID uuid.UUID) (*repositories.SalaryRoleItem, error)
	SaveSalaryRoleDraftStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]repositories.SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error

	// CreateSalaryRole/UpdateSalaryRole are the only code allowed to
	// construct or mutate the domain.SalaryRole entity - see
	// plan_payroll.go's NewSalaryRole.
	CreateSalaryRole(cmd *commands.CreateSalaryRole) (*commands.CreateSalaryRoleResult, error)
	UpdateSalaryRole(cmd *commands.UpdateSalaryRole) (*commands.UpdateSalaryRoleResult, error)

	// Benefits
	CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error)
	FindBenefitDraft(planID uuid.UUID) (*repositories.BenefitItem, error)
	GetBenefit(itemID uuid.UUID) (*repositories.BenefitItem, error)
	SaveBenefitDraftStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int) error
	ListCompleteBenefits(planID uuid.UUID) ([]repositories.BenefitItem, error)
	DeleteBenefit(itemID uuid.UUID) error

	// CreateBenefit/UpdateBenefit are the only code allowed to construct or
	// mutate the domain.Benefit entity - see plan_payroll.go's NewBenefit.
	CreateBenefit(cmd *commands.CreateBenefit) (*commands.CreateBenefitResult, error)
	UpdateBenefit(cmd *commands.UpdateBenefit) (*commands.UpdateBenefitResult, error)

	// Payroll Tax Rates (singleton)
	GetPayrollTaxRatesRow(planID uuid.UUID) (*repositories.PayrollTaxRatesRow, error)
	SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int, isComplete bool) error
}
