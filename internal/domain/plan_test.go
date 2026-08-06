package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// --- Helper Functions ---

func newValidPlan(t *testing.T) *Plan {
	t.Helper()
	plan, err := NewPlan(uuid.New(), "Simple Startup", 1, 2024, uuid.New())
	if err != nil {
		t.Fatalf("failed to create valid plan: %v", err)
	}
	return plan
}

// --- Tests ---

func TestNewPlan(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		testName      string
		id            uuid.UUID
		planName      string
		startingMonth int
		startingYear  int
		expectedErr   error
	}{
		{
			testName:      "Valid business plan",
			id:            validID,
			planName:      "Coffee Shop",
			startingMonth: 1,
			startingYear:  2024,
			expectedErr:   nil,
		},
		{
			testName:      "Fails on empty name",
			id:            uuid.New(),
			planName:      "   ",
			startingMonth: 1,
			startingYear:  2024,
			expectedErr:   ErrInvalidName,
		},
		{
			testName:      "Fails on invalid month (0)",
			id:            uuid.New(),
			planName:      "Mass Market",
			startingMonth: 0,
			startingYear:  2024,
			expectedErr:   ErrInvalidStartingMonth,
		},
		{
			testName:      "Fails on invalid year",
			id:            uuid.New(),
			planName:      "Mass Market",
			startingMonth: 1,
			startingYear:  1800, // Assuming the 1900-2100 bounds from our helper
			expectedErr:   ErrInvalidStartingYear,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			plan, err := NewPlan(tc.id, tc.planName, tc.startingMonth, tc.startingYear, uuid.New())

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil {
				if plan.Name() != tc.planName {
					t.Errorf("expected name %s, got %s", tc.planName, plan.Name())
				}
				if plan.ID() != tc.id {
					t.Errorf("expected id %s, got %s", tc.id, plan.ID())
				}
			}
		})
	}
}
