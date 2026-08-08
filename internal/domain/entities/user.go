package entities

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
)

var (
	ErrInvalidEmail    = errors.New("email cannot be empty")
	ErrInvalidUserName = errors.New("first and last name are required")
	ErrUserNotFound    = errors.New("user not found")
	ErrAccessDenied    = errors.New("access denied")
)

type User struct {
	id        uuid.UUID
	email     string
	firstName string
	lastName  string

	domainEvents []domevents.DomainEvent
}

func (u *User) recordEvent(e domevents.DomainEvent) {
	u.domainEvents = append(u.domainEvents, e)
}

// PullEvents returns all accumulated domain events and resets the list.
// Call this inside the repository's SaveUserWithPassword implementation
// after persisting the user row, in the same transaction as the outbox write.
func (u *User) PullEvents() []domevents.DomainEvent {
	evts := u.domainEvents
	u.domainEvents = nil
	return evts
}

func NewUser(email string) (*User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if cleanEmail == "" {
		return nil, ErrInvalidEmail
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user id: %w", err)
	}

	return &User{
		id:    id,
		email: cleanEmail,
	}, nil
}

func (u *User) ID() uuid.UUID          { return u.id }
func (u *User) Email() string          { return u.email }
func (u *User) FirstName() string      { return u.firstName }
func (u *User) LastName() string       { return u.lastName }
func (u *User) SetID(id uuid.UUID)     { u.id = id }
func (u *User) SetFirstName(fn string) { u.firstName = fn }
func (u *User) SetLastName(ln string)  { u.lastName = ln }

// AggregateID implements entities.AggregateRoot.
//
// User is the second aggregate root in this domain. It owns authentication
// credentials and account identity. Plan accesses User only by ownerID
// (uuid.UUID) — never by embedding *User. Session and PlanAccess are
// entities within the User/Plan boundary respectively and are also
// referenced by ID across aggregate lines.
func (u *User) AggregateID() uuid.UUID { return u.id }

// compile-time assertion that User satisfies the AggregateRoot marker.
var _ AggregateRoot = (*User)(nil)

// FullName returns the user's display name, falling back to their email
// when no name has been set (e.g. legacy accounts created before names
// were required).
func (u *User) FullName() string {
	name := strings.TrimSpace(u.firstName + " " + u.lastName)
	if name == "" {
		return u.email
	}
	return name
}

// AccessLevel defines what a user can do with a plan
type AccessLevel string

const (
	Owner  AccessLevel = "owner"
	Editor AccessLevel = "editor"
	Viewer AccessLevel = "viewer"
)

func (a AccessLevel) IsValid() bool {
	return a == Owner || a == Editor || a == Viewer
}

// PlanAccess represents the relationship between a user and a plan
type PlanAccess struct {
	PlanID      uuid.UUID
	UserID      uuid.UUID
	AccessLevel AccessLevel
	InvitedAt   int64 // Unix timestamp
}

// NewPlanAccess creates a new plan access record
func NewPlanAccess(planID uuid.UUID, userID uuid.UUID, level AccessLevel) (*PlanAccess, error) {
	if !level.IsValid() {
		return nil, errors.New("invalid access level")
	}

	return &PlanAccess{
		PlanID:      planID,
		UserID:      userID,
		AccessLevel: level,
		InvitedAt:   0, // Will be set by store
	}, nil
}

// CanEdit returns true if the user has edit permissions
func (a PlanAccess) CanEdit() bool {
	return a.AccessLevel == Owner || a.AccessLevel == Editor
}

// CanView returns true if the user has view permissions
func (a PlanAccess) CanView() bool {
	return a.CanEdit() || a.AccessLevel == Viewer
}
