package services

import (
	"strings"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type inviteService struct {
	invites repositories.InviteRepository
	plans   repositories.PlanRepository
}

func NewInviteService(invites repositories.InviteRepository, plans repositories.PlanRepository) interfaces.InviteService {
	return &inviteService{invites: invites, plans: plans}
}

// CreateInvite loads the Plan aggregate, constructs the PlanInvite via
// domain.NewPlanInvite, records the resulting UserInvitedToPlan event on the
// Plan, and persists the invite row and the Plan's outbox event atomically
// via PlanRepository.SaveWithInvite - closing both the constructor-bypass
// and the non-atomic-write gaps documented in AGENTS.md's DDD violations
// audit.
func (s *inviteService) CreateInvite(cmd *commands.CreateInvite) (*commands.CreateInviteResult, error) {
	if strings.EqualFold(strings.TrimSpace(cmd.Email), strings.TrimSpace(cmd.InviterEmail)) {
		return nil, domain.ErrSelfInvite
	}

	existing, err := s.invites.GetInvitesForPlan(cmd.PlanID)
	if err != nil {
		return nil, err
	}
	for _, inv := range existing {
		if inv.Status == domain.InvitePending && strings.EqualFold(inv.Email, strings.TrimSpace(cmd.Email)) {
			return nil, domain.ErrDuplicateInvite
		}
	}

	plan, err := s.plans.Get(cmd.PlanID)
	if err != nil {
		return nil, err
	}

	invite, err := domain.NewPlanInvite(cmd.PlanID, cmd.Email, cmd.AccessLevel, cmd.InvitedBy)
	if err != nil {
		return nil, err
	}

	plan.RecordUserInvited(invite.ID, invite.Email, invite.AccessLevel, cmd.InvitedBy)

	validated, err := plan.Validate()
	if err != nil {
		return nil, err
	}

	if err := s.plans.SaveWithInvite(validated, invite); err != nil {
		return nil, err
	}

	return &commands.CreateInviteResult{Result: mapper.NewInviteResultFromEntity(invite)}, nil
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
