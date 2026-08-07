package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type accessService struct {
	access repositories.AccessRepository
}

func NewAccessService(access repositories.AccessRepository) interfaces.AccessService {
	return &accessService{access: access}
}

func (s *accessService) GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error {
	return s.access.GrantAccess(planID, userID, level)
}
func (s *accessService) GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error) {
	return s.access.GetAccess(planID, userID)
}
func (s *accessService) GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error) {
	return s.access.GetPlanAccess(planID)
}
func (s *accessService) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	return s.access.GetUserPlans(userID)
}
