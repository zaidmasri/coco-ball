package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail = errors.New("email cannot be empty")
	ErrUserNotFound = errors.New("user not found")
	ErrAccessDenied = errors.New("access denied")
)

type User struct {
	id    uuid.UUID
	email string
}

func NewUser(email string) (*User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if cleanEmail == "" {
		return nil, ErrInvalidEmail
	}

	return &User{
		id:    uuid.New(),
		email: cleanEmail,
	}, nil
}

func (u *User) ID() uuid.UUID { return u.id }
func (u *User) Email() string { return u.email }

// AccessLevel defines what a user can do with a plan
type AccessLevel string

const (
	Owner    AccessLevel = "owner"
	Editor   AccessLevel = "editor"
	Viewer   AccessLevel = "viewer"
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
