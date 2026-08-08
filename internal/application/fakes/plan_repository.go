package fakes

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
)

// PlanRepository is an in-memory implementation of repositories.PlanRepository
// for use in application-layer tests. It is not safe for concurrent writes
// from multiple goroutines, but is safe for the sequential test patterns
// used in application service tests.
type PlanRepository struct {
	mu     sync.RWMutex
	plans  map[uuid.UUID]*domain.Plan
	events []domevents.DomainEvent
}

func NewPlanRepository() *PlanRepository {
	return &PlanRepository{plans: make(map[uuid.UUID]*domain.Plan)}
}

func (r *PlanRepository) Save(vp domain.ValidatedPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := vp.Plan()
	r.events = append(r.events, p.PullEvents()...)
	r.plans[p.ID()] = p
	return nil
}

func (r *PlanRepository) Get(id uuid.UUID) (*domain.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plans[id]
	if !ok {
		return nil, errors.New("plan not found")
	}
	return p, nil
}

func (r *PlanRepository) GetAll() ([]*domain.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Plan, 0, len(r.plans))
	for _, p := range r.plans {
		out = append(out, p)
	}
	return out, nil
}

func (r *PlanRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.plans[id]; !ok {
		return errors.New("plan not found")
	}
	delete(r.plans, id)
	return nil
}

func (r *PlanRepository) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	return r.GetAll()
}

// DrainEvents returns all domain events recorded since the last call and
// clears the internal buffer. Useful for asserting event emission in tests.
func (r *PlanRepository) DrainEvents() []domevents.DomainEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	evts := r.events
	r.events = nil
	return evts
}
