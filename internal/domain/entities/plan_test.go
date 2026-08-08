package entities

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// --- Helper Functions ---

func newValidPlan(t *testing.T) *Plan {
	t.Helper()
	ownerID, _ := uuid.NewV7()
	plan, err := NewPlan("Simple Startup", 1, 2024, ownerID)
	if err != nil {
		t.Fatalf("failed to create valid plan: %v", err)
	}
	return plan
}

// --- Tests ---

func TestNewPlan(t *testing.T) {
	ownerID, _ := uuid.NewV7()

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
			plan, err := NewPlan(tc.planName, tc.startingMonth, tc.startingYear, ownerID)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil {
				if plan.Name() != tc.planName {
					t.Errorf("expected name %s, got %s", tc.planName, plan.Name())
				}
				if plan.ID() == uuid.Nil {
					t.Errorf("expected a non-nil plan ID")
				}
			}
		})
	}
}
