package interfaces

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type InviteService interface {
	// CreateInvite is the only entry point that may construct a PlanInvite
	// — callers must not call domain.NewPlanInvite themselves.
	CreateInvite(cmd *commands.CreateInvite) (*commands.CreateInviteResult, error)
	GetInvite(id uuid.UUID) (*domain.PlanInvite, error)
	GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error)
	GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error)
	UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error
}
