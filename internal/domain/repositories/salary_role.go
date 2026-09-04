package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type SalaryRoleRepository interface {
	CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error)
	FindSalaryRoleDraft(planID uuid.UUID) (*SalaryRoleItem, error)
	GetSalaryRole(itemID uuid.UUID) (*SalaryRoleItem, error)
	// SaveSalaryRoleDraftStep persists an in-progress, possibly incomplete
	// wizard step. Drafts are allowed to be invalid by design, so this
	// accepts a bare domain.SalaryRole.
	SaveSalaryRoleDraftStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int) error
	// CompleteSalaryRole marks a salary role item complete. It requires a
	// domain.ValidatedSalaryRole, so it is a compile error to complete an
	// item without validating it first - see domain.NewValidatedSalaryRole.
	CompleteSalaryRole(itemID uuid.UUID, role domain.ValidatedSalaryRole, currentStep int) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error
}
