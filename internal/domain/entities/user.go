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

// VerifiedUser is an opaque token proving a uuid.UUID names a real, existing
// User of the system — not just any uuid.UUID a caller happens to supply.
// It can only be produced by wrapping a *User a repository lookup actually
// returned (see UserRepository.GetUser); there is no constructor that
// accepts a bare uuid.UUID. Plan.NewPlan requires a VerifiedUser as its
// owner parameter, so a plan can never be created for a user that hasn't
// been confirmed to exist — the same "invalid states unrepresentable"
// pattern as ValidatedPlan (plan.go), mirroring go-ddd's ValidatedSeller
// (a Product can only be created against a Seller that passed validation).
type VerifiedUser struct {
	user *User
}

// NewVerifiedUser wraps a User the caller has already confirmed exists — in
// practice, a *User returned by UserRepository.GetUser. It does not perform
// the existence check itself: the domain layer has no database access, so
// that check belongs to the application service calling this (see
// PlanService.CreatePlan).
func NewVerifiedUser(u *User) (VerifiedUser, error) {
	if u == nil {
		return VerifiedUser{}, ErrUserNotFound
	}
	return VerifiedUser{user: u}, nil
}

// ID returns the verified user's identity.
func (vu VerifiedUser) ID() uuid.UUID { return vu.user.ID() }

// User returns the wrapped user.
func (vu VerifiedUser) User() *User { return vu.user }

// ChangeName updates the user's first and last name, validating both are
// non-empty after trimming. Mirrors Plan.ChangeCoreDetails: validate, then
// mutate only after validation passes, then record the domain event. This is
// the validated entry point for a name change — SetFirstName/SetLastName
// remain reconstruction-only (no validation), used by UserRepository.GetUser.
func (u *User) ChangeName(firstName, lastName string) error {
	cleanFirst := strings.TrimSpace(firstName)
	cleanLast := strings.TrimSpace(lastName)
	if cleanFirst == "" || cleanLast == "" {
		return ErrInvalidUserName
	}

	u.firstName = cleanFirst
	u.lastName = cleanLast

	u.recordEvent(domevents.NewUserUpdated(u.id))
	return nil
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
