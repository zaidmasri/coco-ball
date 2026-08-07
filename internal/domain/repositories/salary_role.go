package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type SalaryRoleRepository interface {
	CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error)
	GetSalaryRoleDraft(planID uuid.UUID) (*SalaryRoleItem, error)
	GetSalaryRole(itemID uuid.UUID) (*SalaryRoleItem, error)
	SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status ItemStatus) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error
}
