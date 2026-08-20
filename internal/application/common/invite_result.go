package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// InviteResult is the write-side acknowledgement shape for CreateInvite. It
// deliberately does not carry the full PlanInvite's mutation surface —
// callers of CreateInvite only need enough to redirect and display the
// pending invite.
type InviteResult struct {
	ID          uuid.UUID
	PlanID      uuid.UUID
	Email       string
	AccessLevel domain.AccessLevel
	Status      domain.InviteStatus
	InvitedBy   uuid.UUID
}
