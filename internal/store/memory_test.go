package store

import (
	"sync"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

func TestMemoryStore_GenerateID(t *testing.T) {
	store := NewMemoryStore()

	id1 := store.GenerateID()
	id2 := store.GenerateID()

	if id1 != 1 {
		t.Errorf("expected first ID to be 1, got %d", id1)
	}
	if id2 != 2 {
		t.Errorf("expected second ID to be 2, got %d", id2)
	}
}

func TestMemoryStore_SaveAndGet(t *testing.T) {
	store := NewMemoryStore()
	plan, _ := domain.NewPlan(1, "Test Business")

	// 1. Test Get on an empty store
	_, err := store.Get(1)
	if err == nil || err.Error() != "plan not found" {
		t.Errorf("expected 'plan not found' error, got %v", err)
	}

	// 2. Test Save
	err = store.Save(plan)
	if err != nil {
		t.Errorf("expected no error on save, got %v", err)
	}

	// 3. Test Get on a populated store
	savedPlan, err := store.Get(1)
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
	plan1, _ := domain.NewPlan(1, "Plan One")
	plan2, _ := domain.NewPlan(2, "Plan Two")
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
		go func(id int) {
			defer wg.Done()
			plan, _ := domain.NewPlan(id, "Concurrent Plan")
			_ = store.Save(plan)
		}(i)
	}
	wg.Wait() // Wait for all writes to finish

	// Verify all writes succeeded
	plans, _ := store.GetAll()
	if len(plans) != workers {
		t.Fatalf("expected %d plans saved, got %d", workers, len(plans))
	}

	// 2. Concurrent Reads and ID Generations (Simulating a mix of POST and GET requests)
	for i := 1; i <= workers; i++ {
		wg.Add(2) // Adding 2 because we are launching 2 goroutines per worker

		// Routine A: Try to read the plan
		go func(id int) {
			defer wg.Done()
			_, _ = store.Get(id)
		}(i)

		// Routine B: Try to generate a new ID
		go func() {
			defer wg.Done()
			_ = store.GenerateID()
		}()
	}
	wg.Wait() // Wait for all reads and ID generations to finish
}
