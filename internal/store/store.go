// Package store handles persistence of system data
package store

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// PlanStore defines how we save and retrieve plans.
type PlanStore interface {
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
	Delete(id uuid.UUID) error

	// User management (basic)
	SaveUser(u *domain.User) error
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)

	// User management (with password)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)

	// Session management
	SaveSession(s *domain.Session) error
	GetSession(sessionID string) (*domain.Session, error)
	DeleteSession(sessionID string) error

	// Access control
	GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error
	GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error)
	GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error)
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)
}
