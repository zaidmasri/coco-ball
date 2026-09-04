package entities

import (
	"errors"
	"strings"

	"uuid"
)

var (
	ErrInviteNotFound   = errors.New("invite not found")
	ErrInviteNotPending = errors.New("invite has already been responded to")
	ErrInviteForbidden  = errors.New("invite does not belong to this user")
	ErrSelfInvite       = errors.New("you cannot invite yourself to your own plan")
	ErrDuplicateInvite  = errors.New("this email already has a pending invite for this plan")
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

// validate checks a PlanInvite's business invariants (mirrors Cost.validate()).
func (pi *PlanInvite) validate() error {
	if pi.Email == "" {
		return ErrInvalidEmail
	}
	if err := validateEmailFormat(pi.Email); err != nil {
		return err
	}
	if !pi.AccessLevel.IsValid() {
		return errors.New("invalid access level")
	}
	return nil
}

// NewPlanInvite creates a new pending invite for a plan.
func NewPlanInvite(planID uuid.UUID, email string, level AccessLevel, invitedBy uuid.UUID) (*PlanInvite, error) {
	id := uuid.NewV7()

	invite := &PlanInvite{
		ID:          id,
		PlanID:      planID,
		Email:       strings.TrimSpace(strings.ToLower(email)),
		AccessLevel: level,
		Status:      InvitePending,
		InvitedBy:   invitedBy,
	}
	if err := invite.validate(); err != nil {
		return nil, err
	}
	return invite, nil
}

// ValidatedPlanInvite is an opaque token proving a PlanInvite passed every
// invariant PlanInvite.validate() checks. It can only be produced by
// NewValidatedPlanInvite - mirrors ValidatedPlan's shape (plan.go).
// PlanRepository.SaveWithInvite accepts only this type, not a bare
// *PlanInvite, so an invite can never be persisted without validation.
type ValidatedPlanInvite struct {
	invite      *PlanInvite
	isValidated bool
}

// NewValidatedPlanInvite validates an existing *PlanInvite and wraps it.
func NewValidatedPlanInvite(invite *PlanInvite) (ValidatedPlanInvite, error) {
	if err := invite.validate(); err != nil {
		return ValidatedPlanInvite{}, err
	}
	return ValidatedPlanInvite{invite: invite, isValidated: true}, nil
}

func (v ValidatedPlanInvite) Invite() *PlanInvite { return v.invite }
