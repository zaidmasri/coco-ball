// Package store handles persistance of system data
package store

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// PlanStore defines how we save and retrieve plans.
type PlanStore interface {
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
}

// MemoryStore is our temporary database.
type MemoryStore struct {
	mu    sync.RWMutex
	plans map[uuid.UUID]*domain.Plan
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		plans: make(map[uuid.UUID]*domain.Plan),
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
