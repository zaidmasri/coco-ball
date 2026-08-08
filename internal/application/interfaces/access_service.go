package interfaces

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type AccessService interface {
	GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error
	GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error)
	GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error)
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)
}
