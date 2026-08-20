package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// InviteRepository is a sub-entity repository for PlanInvite, which lives
// within the Plan aggregate boundary (see entities/invite.go's doc
// comment). Creating an invite is not exposed here — PlanRepository.
// SaveWithInvite persists the invite row atomically with the Plan
// aggregate's own save and its UserInvitedToPlan outbox event, since a
// separate, non-transactional CreateInvite call could silently lose that
// event. This interface covers reads and status updates only.
type InviteRepository interface {
	GetInvite(id uuid.UUID) (*domain.PlanInvite, error)
	GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error)
	GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error)
	UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error
}
