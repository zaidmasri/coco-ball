package events

import "github.com/google/uuid"

const UserRegisteredEventName = "user.registered"

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
