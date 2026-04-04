package domain

import (
	"errors"
	"testing"
)

// --- Helper Functions ---

func newValidPlan(t *testing.T) *Plan {
	t.Helper()
	plan, err := NewPlan(1, "Simple Startup")
	if err != nil {
		t.Fatalf("failed to create valid plan: %v", err)
	}
	return plan
}

// --- Tests ---

func TestNewPlan(t *testing.T) {
	tests := []struct {
		testName    string
		id          int
		planName    string
		expectedErr error
	}{
		{
			testName:    "Valid business plan",
			id:          1,
			planName:    "Coffee Shop",
			expectedErr: nil,
		},
		{
			testName:    "Fails on empty name",
			id:          2,
			planName:    "   ",
			expectedErr: ErrInvalidName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			plan, err := NewPlan(tc.id, tc.planName)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil {
				if plan.Name() != tc.planName {
					t.Errorf("expected name %s, got %s", tc.planName, plan.Name())
				}
				if plan.ID() != tc.id {
					t.Errorf("expected id %d, got %d", tc.id, plan.ID())
				}
			}
		})
	}
}

func TestPlan_AddExpense(t *testing.T) {
	plan := newValidPlan(t)

	t.Run("Valid Expense", func(t *testing.T) {
		err := plan.AddExpense("Rent", 5000)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(plan.Expenses()) != 1 {
			t.Errorf("expected 1 expense, got %d", len(plan.Expenses()))
		}
	})

	t.Run("Fails on Negative Amount", func(t *testing.T) {
		err := plan.AddExpense("Refunds", -100)
		if !errors.Is(err, ErrNegativeAmount) {
			t.Errorf("expected ErrNegativeAmount, got %v", err)
		}
	})

	t.Run("Fails on Empty Name", func(t *testing.T) {
		err := plan.AddExpense(" ", 500)
		if !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
	})
}

func TestPlan_TotalExpenses(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddExpense("Rent", 2000)
	_ = plan.AddExpense("Software", 500)
	_ = plan.AddExpense("Insurance", 300)

	expectedTotal := Money(2800)

	if got := plan.TotalExpenses(); got != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, got)
	}
}
