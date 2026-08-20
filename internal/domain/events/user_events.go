package events

import "uuid"

const (
	UserRegisteredEventName = "user.registered"
	UserUpdatedEventName    = "user.updated"
)

// UserRegistered is emitted by the User aggregate root when a new account
// is created via NewUserWithPassword.
type UserRegistered struct {
	BaseEvent
	Email string
}

func NewUserRegistered(userID uuid.UUID, email string) UserRegistered {
	return UserRegistered{
		BaseEvent: NewBaseEvent(userID),
		Email:     email,
	}
}

func (e UserRegistered) EventName() string { return UserRegisteredEventName }

// UserUpdated is emitted when a user's profile details change.
type UserUpdated struct {
	BaseEvent
}

func NewUserUpdated(userID uuid.UUID) UserUpdated {
	return UserUpdated{BaseEvent: NewBaseEvent(userID)}
}

func (e UserUpdated) EventName() string { return UserUpdatedEventName }
