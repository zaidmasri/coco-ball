package entities

import (
	"errors"
	"fmt"
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
//
// PlanInvite is an entity within the Plan aggregate boundary. It does not
// carry an event buffer — only aggregate roots may emit events. The Plan
// aggregate emits UserInvitedToPlan via Plan.RecordUserInvited after the
// invite is created.
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

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite id: %w", err)
	}

	return &PlanInvite{
		ID:          id,
		PlanID:      planID,
		Email:       cleanEmail,
		AccessLevel: level,
		Status:      InvitePending,
		InvitedBy:   invitedBy,
	}, nil
}
