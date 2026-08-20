package services

import (
	"errors"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type accessService struct {
	access repositories.AccessRepository
	plans  repositories.PlanRepository
	users  repositories.UserRepository
}

func NewAccessService(access repositories.AccessRepository, plans repositories.PlanRepository, users repositories.UserRepository) interfaces.AccessService {
	return &accessService{access: access, plans: plans, users: users}
}

// GrantAccess verifies planID/userID both name real, existing rows before
// delegating the insert to the repository - this existence-check business
// logic used to live in AccessRepository.GrantAccess (raw COUNT queries),
// violating "no business logic in internal/infrastructure/sqlite/".
func (s *accessService) GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error {
	if !level.IsValid() {
		return errors.New("invalid access level")
	}
	if _, err := s.plans.Get(planID); err != nil {
		return err
	}
	if _, err := s.users.GetUser(userID); err != nil {
		return err
	}
	return s.access.GrantAccess(planID, userID, level)
}
func (s *accessService) GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error) {
	return s.access.GetAccess(planID, userID)
}
func (s *accessService) GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error) {
	return s.access.GetPlanAccess(planID)
}
