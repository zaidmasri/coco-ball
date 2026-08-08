package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type SalaryRoleRepository interface {
	CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error)
	FindSalaryRoleDraft(planID uuid.UUID) (*SalaryRoleItem, error)
	GetSalaryRole(itemID uuid.UUID) (*SalaryRoleItem, error)
	SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status ItemStatus) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error
}
