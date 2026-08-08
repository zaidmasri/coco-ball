package interfaces

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type InviteService interface {
	CreateInvite(invite *domain.PlanInvite) error
	GetInvite(id uuid.UUID) (*domain.PlanInvite, error)
	GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error)
	GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error)
	UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error
}
