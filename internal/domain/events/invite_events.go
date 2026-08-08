package events

import "github.com/google/uuid"

const UserInvitedToPlanEventName = "plan.user_invited"

// UserInvitedToPlan is emitted by the Plan aggregate root when a collaborator
// invite is created. PlanInvite is not an aggregate root and must not emit
// events directly; the Plan emits this event via RecordUserInvited.
type UserInvitedToPlan struct {
	BaseEvent
	InviteID    uuid.UUID
	InvitedBy   uuid.UUID
	Email       string
	AccessLevel string
}

func NewUserInvitedToPlan(planID, inviteID, invitedBy uuid.UUID, email, accessLevel string) UserInvitedToPlan {
	return UserInvitedToPlan{
		BaseEvent:   NewBaseEvent(planID),
		InviteID:    inviteID,
		InvitedBy:   invitedBy,
		Email:       email,
		AccessLevel: accessLevel,
	}
}

func (e UserInvitedToPlan) EventName() string { return UserInvitedToPlanEventName }
