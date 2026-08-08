package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type inviteService struct {
	invites repositories.InviteRepository
}

func NewInviteService(invites repositories.InviteRepository) interfaces.InviteService {
	return &inviteService{invites: invites}
}

func (s *inviteService) CreateInvite(invite *domain.PlanInvite) error {
	return s.invites.CreateInvite(invite)
}
func (s *inviteService) GetInvite(id uuid.UUID) (*domain.PlanInvite, error) {
	return s.invites.GetInvite(id)
}
func (s *inviteService) GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error) {
	return s.invites.GetInvitesForPlan(planID)
}
func (s *inviteService) GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error) {
	return s.invites.GetPendingInvitesForEmail(email)
}
func (s *inviteService) UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error {
	return s.invites.UpdateInviteStatus(id, status)
}
