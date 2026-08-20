package entities

import (
	"errors"
	"testing"

	"uuid"
)

// --- Helper Functions ---

// newVerifiedOwner builds a VerifiedUser for domain-layer tests. NewPlan
// only requires proof of verification, not an actual database row, so
// wrapping a freshly-constructed User is sufficient here — the real
// existence check happens in PlanService.CreatePlan (see
// plan_service_test.go's TestPlanService_CreatePlan_RejectsUnknownOwner).
func newVerifiedOwner(t *testing.T) VerifiedUser {
	t.Helper()
	user, err := NewUser("owner@example.com")
	if err != nil {
		t.Fatalf("failed to create owner user: %v", err)
	}
	verified, err := NewVerifiedUser(user)
	if err != nil {
		t.Fatalf("failed to verify owner user: %v", err)
	}
	return verified
}

func newValidPlan(t *testing.T) *Plan {
	t.Helper()
	plan, err := NewPlan("Simple Startup", 1, 2024, newVerifiedOwner(t))
	if err != nil {
		t.Fatalf("failed to create valid plan: %v", err)
	}
	return plan
}

// --- Tests ---

func TestNewPlan(t *testing.T) {
	owner := newVerifiedOwner(t)

	tests := []struct {
		testName      string
		planName      string
		startingMonth int
		startingYear  int
		expectedErr   error
	}{
		{
			testName:      "Valid business plan",
			planName:      "Coffee Shop",
			startingMonth: 1,
			startingYear:  2024,
			expectedErr:   nil,
		},
		{
			testName:      "Fails on empty name",
			planName:      "   ",
			startingMonth: 1,
			startingYear:  2024,
			expectedErr:   ErrInvalidName,
		},
		{
			testName:      "Fails on invalid month (0)",
			planName:      "Mass Market",
			startingMonth: 0,
			startingYear:  2024,
			expectedErr:   ErrInvalidStartingMonth,
		},
		{
			testName:      "Fails on invalid year",
			planName:      "Mass Market",
			startingMonth: 1,
			startingYear:  1800,
			expectedErr:   ErrInvalidStartingYear,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			plan, err := NewPlan(tc.planName, tc.startingMonth, tc.startingYear, owner)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil {
				if plan.Name() != tc.planName {
					t.Errorf("expected name %s, got %s", tc.planName, plan.Name())
				}
				if plan.ID() == uuid.Nil() {
					t.Errorf("expected a non-nil plan ID")
				}
			}
		})
	}
}

func TestPlan_Delete_EmitsPlanDeletedEvent(t *testing.T) {
	plan := newValidPlan(t)
	plan.PullEvents() // drain the PlanCreated event from NewPlan first

	plan.Delete()

	events := plan.PullEvents()
	if len(events) != 1 {
		t.Fatalf("expected exactly 1 event after Delete, got %d", len(events))
	}
	if events[0].EventName() != "plan.deleted" {
		t.Errorf("expected plan.deleted event, got %q", events[0].EventName())
	}
	if events[0].AggregateID() != plan.ID() {
		t.Errorf("expected event aggregate id %v, got %v", plan.ID(), events[0].AggregateID())
	}
}
