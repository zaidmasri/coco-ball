package domain

import (
	"errors"
	"testing"
)

// --- Helper Functions ---

func newValidPlan(t *testing.T) *Plan {
	t.Helper()
	plan, err := NewPlan(1, "Simple Startup", 12)
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
		duration    int
		expectedErr error
	}{
		{
			testName:    "Valid business plan",
			id:          1,
			planName:    "Coffee Shop",
			expectedErr: nil,
			duration:    12,
		},
		{
			testName:    "Fails on empty name",
			id:          2,
			planName:    "   ",
			expectedErr: ErrInvalidName,
			duration:    12,
		},
		{
			testName:    "Fails on zero duration",
			id:          3,
			planName:    "Mass Market",
			expectedErr: ErrInvalidDuration,
			duration:    0,
		},

		{
			testName:    "Fails on negative duration",
			id:          4,
			planName:    "Mass Market",
			expectedErr: ErrInvalidDuration,
			duration:    0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			plan, err := NewPlan(tc.id, tc.planName, tc.duration)

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
		err := plan.AddExpense("Rent", 5000, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(plan.Expenses()) != 1 {
			t.Errorf("expected 1 expense, got %d", len(plan.Expenses()))
		}
	})

	t.Run("Fails on Negative Amount", func(t *testing.T) {
		err := plan.AddExpense("Refunds", -100, 0)
		if !errors.Is(err, ErrNegativeAmount) {
			t.Errorf("expected ErrNegativeAmount, got %v", err)
		}
	})

	t.Run("Fails on Empty Name", func(t *testing.T) {
		err := plan.AddExpense(" ", 500, 0)
		if !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
	})
}

func TestPlan_TotalExpenses(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddExpense("Rent", 2000, 0)
	_ = plan.AddExpense("Software", 500, 0)
	_ = plan.AddExpense("Insurance", 300, 0)

	expectedTotal := Money(2800)

	if got := plan.TotalExpenses(); got != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, got)
	}
}

func TestPlan_MonthIndexBounds(t *testing.T) {
	plan := newValidPlan(t)

	// Valid Month (Month 0 is the 1st month)
	err := plan.AddExpense("Valid Rent", 1000, 0)
	if err != nil {
		t.Errorf("expected no error for MonthIndex 0, got %v", err)
	}

	// Valid End Month (Month 11 is the 12th month)
	err = plan.AddRevenue("Valid Sales", 5000, 11)
	if err != nil {
		t.Errorf("expected no error for MonthIndex 11, got %v", err)
	}

	// Invalid Negative Month
	err = plan.AddExpense("Time Travel", 100, -1)
	if !errors.Is(err, ErrInvalidMonthIndex) {
		t.Errorf("expected ErrInvalidMonthIndex for -1, got %v", err)
	}

	// Invalid Future Month
	err = plan.AddRevenue("Distant Future", 100, 12)
	if !errors.Is(err, ErrInvalidMonthIndex) {
		t.Errorf("expected ErrInvalidMonthIndex for 12, got %v", err)
	}
}

func TestPlan_MonthlyLedgerMath(t *testing.T) {
	plan := newValidPlan(t)

	// Month 0 setup (Loss)
	_ = plan.AddRevenue("Sales", 2000, 0)
	_ = plan.AddExpense("Rent", 1500, 0)
	_ = plan.AddExpense("Marketing", 1000, 0) // Total Expense: 2500

	// Month 1 setup (Profit)
	_ = plan.AddRevenue("Sales", 5000, 1)
	_ = plan.AddExpense("Rent", 1500, 1) // Total Expense: 1500

	// Assertions Month 0
	if got := plan.MonthlyNetCashFlow(0); got != -500 {
		t.Errorf("expected Month 0 net flow of -500, got %d", got)
	}

	// Assertions Month 1
	if got := plan.MonthlyNetCashFlow(1); got != 3500 {
		t.Errorf("expected Month 1 net flow of 3500, got %d", got)
	}

	// Assertions Month 2 (Empty)
	if got := plan.MonthlyNetCashFlow(2); got != 0 {
		t.Errorf("expected Month 2 net flow of 0, got %d", got)
	}

	// Assertions Total Expense
	if got := plan.TotalExpenses(); got != 4000 {
		t.Errorf("expected Total Expenses of 4000, got %d", got)
	}

	// Assertions Total Revenues
	if got := plan.TotalRevenues(); got != 7000 {
		t.Errorf("expected Total Expenses of 7000, got %d", got)
	}
}
