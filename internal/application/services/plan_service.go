package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type planService struct {
	plans repositories.PlanRepository
}

func NewPlanService(plans repositories.PlanRepository) interfaces.PlanService {
	return &planService{plans: plans}
}

func (s *planService) Save(p *domain.Plan) error          { return s.plans.Save(p) }
func (s *planService) Get(id uuid.UUID) (*domain.Plan, error) { return s.plans.Get(id) }
func (s *planService) GetAll() ([]*domain.Plan, error)    { return s.plans.GetAll() }
func (s *planService) Delete(id uuid.UUID) error          { return s.plans.Delete(id) }
func (s *planService) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	return s.plans.GetUserPlans(userID)
}
