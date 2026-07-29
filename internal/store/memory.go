// Package store handles persistance of system data
package store

import (
	"errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// PlanStore defines how we save and retrieve plans.
type PlanStore interface {
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)

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

// MemoryStore is our temporary database.
type MemoryStore struct {
	mu               sync.RWMutex
	plans            map[uuid.UUID]*domain.Plan
	users            map[uuid.UUID]*domain.User
	userCreds        map[string]*domain.UserWithPassword // email -> UserWithPassword
	emails           map[string]uuid.UUID                // email -> userID
	sessions         map[string]*domain.Session          // sessionID -> Session
	access           map[string]*domain.PlanAccess       // "planID:userID" -> PlanAccess
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		plans:     make(map[uuid.UUID]*domain.Plan),
		users:     make(map[uuid.UUID]*domain.User),
		userCreds: make(map[string]*domain.UserWithPassword),
		emails:    make(map[string]uuid.UUID),
		sessions:  make(map[string]*domain.Session),
		access:    make(map[string]*domain.PlanAccess),
	}
}

func (m *MemoryStore) GetAll() ([]*domain.Plan, error) {
	m.mu.RLock() // Protect against concurrent map reads
	defer m.mu.RUnlock()

	// 1. Create a slice with a capacity equal to the map's length.
	// (Defining the capacity upfront is a Go best practice for performance).
	plans := make([]*domain.Plan, 0, len(m.plans))

	// 2. Iterate over the map and append just the values (the plans) to the slice.
	for _, plan := range m.plans {
		plans = append(plans, plan)
	}

	return plans, nil
}

func (m *MemoryStore) Save(p *domain.Plan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plans[p.ID()] = p
	return nil
}

func (m *MemoryStore) Get(id uuid.UUID) (*domain.Plan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	plan, exists := m.plans[id]
	if !exists {
		return nil, errors.New("plan not found")
	}
	return plan, nil
}

// SaveUser stores a user in memory
func (m *MemoryStore) SaveUser(u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[u.ID()] = u
	m.emails[u.Email()] = u.ID()
	return nil
}

// GetUser retrieves a user by ID
func (m *MemoryStore) GetUser(id uuid.UUID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email address
func (m *MemoryStore) GetUserByEmail(email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID, exists := m.emails[strings.ToLower(email)]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	user, ok := m.users[userID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GrantAccess gives a user access to a plan
func (m *MemoryStore) GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !level.IsValid() {
		return errors.New("invalid access level")
	}

	// Verify plan and user exist
	if _, exists := m.plans[planID]; !exists {
		return errors.New("plan not found")
	}
	if _, exists := m.users[userID]; !exists {
		return domain.ErrUserNotFound
	}

	key := planID.String() + ":" + userID.String()
	m.access[key] = &domain.PlanAccess{
		PlanID:      planID,
		UserID:      userID,
		AccessLevel: level,
		InvitedAt:   int64(0), // Use current time in production
	}
	return nil
}

// GetAccess retrieves access information for a specific user on a specific plan
func (m *MemoryStore) GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := planID.String() + ":" + userID.String()
	access, exists := m.access[key]
	if !exists {
		return nil, domain.ErrAccessDenied
	}
	return access, nil
}

// GetPlanAccess retrieves all access records for a plan
func (m *MemoryStore) GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var accesses []*domain.PlanAccess
	for _, access := range m.access {
		if access.PlanID == planID {
			accesses = append(accesses, access)
		}
	}
	return accesses, nil
}

// GetUserPlans retrieves all plans a user has access to
func (m *MemoryStore) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plans []*domain.Plan
	for _, access := range m.access {
		if access.UserID == userID {
			if plan, exists := m.plans[access.PlanID]; exists {
				plans = append(plans, plan)
			}
		}
	}
	return plans, nil
}

// SaveUserWithPassword stores a user with password credentials
func (m *MemoryStore) SaveUserWithPassword(u *domain.UserWithPassword) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Save basic user info
	m.users[u.ID()] = u.User
	m.emails[u.Email()] = u.ID()

	// Save credentials
	m.userCreds[u.Email()] = u

	return nil
}

// GetUserWithPassword retrieves a user with their password hash by email
func (m *MemoryStore) GetUserWithPassword(email string) (*domain.UserWithPassword, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userCreds, exists := m.userCreds[strings.ToLower(email)]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return userCreds, nil
}

// SaveSession stores a session
func (m *MemoryStore) SaveSession(s *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[s.ID] = s
	return nil
}

// GetSession retrieves a session by ID
func (m *MemoryStore) GetSession(sessionID string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, domain.ErrSessionNotFound
	}

	if !session.IsValid() {
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

// DeleteSession removes a session
func (m *MemoryStore) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}
