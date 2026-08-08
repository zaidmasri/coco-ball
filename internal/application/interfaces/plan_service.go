package interfaces

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type PlanService interface {
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
	Delete(id uuid.UUID) error
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)
}
