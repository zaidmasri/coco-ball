package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type AccessRepository interface {
	GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error
	GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error)
	GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error)
}
