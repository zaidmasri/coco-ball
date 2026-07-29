package store

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

func TestMemoryStore_SaveAndGet(t *testing.T) {
	store := NewMemoryStore()
	id := uuid.New()
	plan, _ := domain.NewPlan(id, "Test Business", 1, 2026, uuid.New())

	// 1. Test Get on an empty store
	_, err := store.Get(id)
	if err == nil || err.Error() != "plan not found" {
		t.Errorf("expected 'plan not found' error, got %v", err)
	}

	// 2. Test Save
	err = store.Save(plan)
	if err != nil {
		t.Errorf("expected no error on save, got %v", err)
	}

	// 3. Test Get on a populated store
	savedPlan, err := store.Get(id)
	if err != nil {
		t.Errorf("expected no error on get, got %v", err)
	}
	if savedPlan.Name() != "Test Business" {
		t.Errorf("expected name 'Test Business', got %s", savedPlan.Name())
	}
}

func TestMemoryStore_GetAll(t *testing.T) {
	store := NewMemoryStore()

	// 1. Test GetAll on empty store
	emptyPlans, err := store.GetAll()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(emptyPlans) != 0 {
		t.Errorf("expected 0 plans, got %d", len(emptyPlans))
	}

	// 2. Populate store
	plan1, _ := domain.NewPlan(uuid.New(), "Plan One", 12, 2025, uuid.New())
	plan2, _ := domain.NewPlan(uuid.New(), "Plan Two", 6, 2026, uuid.New())
	_ = store.Save(plan1)
	_ = store.Save(plan2)

	// 3. Test GetAll on populated store
	plans, err := store.GetAll()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(plans))
	}
}

// TestMemoryStore_Concurrency is the most important test here.
// It spins up 100 goroutines that all attempt to read and write to the
// store at the exact same time to ensure our mutex locks are working correctly.
func TestMemoryStore_Concurrency(t *testing.T) {
	store := NewMemoryStore()
	var wg sync.WaitGroup
	workers := 100

	// 1. Concurrent Writes (Simulating 100 simultaneous POST requests)
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			plan, _ := domain.NewPlan(id, "Concurrent Plan", 2, 2024, uuid.New())
			_ = store.Save(plan)
		}(uuid.New())
	}
	wg.Wait() // Wait for all writes to finish

	// Verify all writes succeeded
	plans, _ := store.GetAll()
	if len(plans) != workers {
		t.Fatalf("expected %d plans saved, got %d", workers, len(plans))
	}

	// 2. Concurrent Reads and Writes (Simulating a mix of POST and GET requests)
	for i := 0; i < workers; i++ {
		wg.Add(2) // Adding 2 because we are launching 2 goroutines per worker

		// Routine A: Try to read the plan using an actual UUID from our saved plans
		go func(p *domain.Plan) {
			defer wg.Done()
			_, _ = store.Get(p.ID())
		}(plans[i]) // Pass the plan from the slice we just verified

		// Routine B: Write brand new plans concurrently to test Lock contention
		go func(id uuid.UUID) {
			defer wg.Done()
			newPlan, _ := domain.NewPlan(id, "Another Concurrent Plan", 3, 2025, uuid.New())
			_ = store.Save(newPlan)
		}(uuid.New())
	}
	wg.Wait() // Wait for all reads and writes to finish
}
