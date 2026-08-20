package interfaces

import (
	"uuid"

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
	SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status repositories.ItemStatus) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]repositories.SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error

	// Benefits
	CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error)
	FindBenefitDraft(planID uuid.UUID) (*repositories.BenefitItem, error)
	GetBenefit(itemID uuid.UUID) (*repositories.BenefitItem, error)
	SaveBenefitStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int, status repositories.ItemStatus) error
	ListCompleteBenefits(planID uuid.UUID) ([]repositories.BenefitItem, error)
	DeleteBenefit(itemID uuid.UUID) error

	// Payroll Tax Rates (singleton)
	GetPayrollTaxRatesRow(planID uuid.UUID) (*repositories.PayrollTaxRatesRow, error)
	SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error
}
