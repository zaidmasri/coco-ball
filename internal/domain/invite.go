package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInviteNotFound   = errors.New("invite not found")
	ErrInviteNotPending = errors.New("invite has already been responded to")
	ErrInviteForbidden  = errors.New("invite does not belong to this user")
)

// InviteStatus tracks whether an invite is awaiting a response.
type InviteStatus string

const (
	InvitePending  InviteStatus = "pending"
	InviteAccepted InviteStatus = "accepted"
	InviteRejected InviteStatus = "rejected"
)

// PlanInvite represents an invitation for an email address to collaborate
// on a plan at a given access level.
type PlanInvite struct {
	ID          uuid.UUID
	PlanID      uuid.UUID
	Email       string
	AccessLevel AccessLevel
	Status      InviteStatus
	InvitedBy   uuid.UUID
	CreatedAt   int64 // Unix timestamp, set by store
	RespondedAt int64 // Unix timestamp, 0 until responded
}

// NewPlanInvite creates a new pending invite for a plan.
func NewPlanInvite(planID uuid.UUID, email string, level AccessLevel, invitedBy uuid.UUID) (*PlanInvite, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if cleanEmail == "" {
		return nil, ErrInvalidEmail
	}
	if !level.IsValid() {
		return nil, errors.New("invalid access level")
	}

	return &PlanInvite{
		ID:          uuid.New(),
		PlanID:      planID,
		Email:       cleanEmail,
		AccessLevel: level,
		Status:      InvitePending,
		InvitedBy:   invitedBy,
	}, nil
}
