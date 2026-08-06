package domain

import (
	"errors"
	"testing"
)

func TestPlan_TotalExpenses(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddCOGS("Metal", 2000, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Software", 500, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddCOGS("Wood", 300, GrowthStrategy{Type: FlatGrowth})

	expectedTotal := Money(33600)

	// FIX: Pass duration (e.g., 12 months) into TotalExpenses
	if got := plan.TotalExpenses(12); got != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, got)
	}
}

func TestPlan_MonthlyLedgerMath(t *testing.T) {
	plan := newValidPlan(t)

	// 5000/mo * 12 months = 60,000
	_ = plan.AddRevenue("Software Subs", 5000, GrowthStrategy{Type: FlatGrowth})

	// 2500/mo * 12 months = 30,000
	_ = plan.AddOpEx("Rent", 1500, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Marketing", 1000, GrowthStrategy{Type: FlatGrowth})

	// --- Monthly Assertions ---

	// FIX: Pass duration (12) to MonthlyNetCashFlow checks
	if got := plan.MonthlyNetCashFlow(0, 12); got != 2500 {
		t.Errorf("expected Month 0 net flow of 2500, got %d", got)
	}
	if got := plan.MonthlyNetCashFlow(11, 12); got != 2500 {
		t.Errorf("expected Month 11 net flow of 2500, got %d", got)
	}

	// Month 12 is out of bounds (12-month plan ends at index 11)
	if got := plan.MonthlyNetCashFlow(12, 12); got != 0 {
		t.Errorf("expected Month 12 net flow of 0 (out of bounds), got %d", got)
	}

	// --- Lifetime Assertions ---

	// FIX: Pass duration (12) to Total methods
	if got := plan.TotalExpenses(12); got != 30000 {
		t.Errorf("expected Total Expenses of 30000, got %d", got)
	}

	if got := plan.TotalRevenues(12); got != 60000 {
		t.Errorf("expected Total Revenues of 60000, got %d", got)
	}
}

func TestCost_ProjectedAmount_Flat(t *testing.T) {
	cost := Cost{
		Name:               "Rent",
		BaseAmountPerMonth: 5000,
		Growth: GrowthStrategy{
			Type: FlatGrowth,
		},
	}

	if got := cost.ProjectedAmount(0); got != 5000 {
		t.Errorf("expected Month 0 to be 5000, got %d", got)
	}
	if got := cost.ProjectedAmount(24); got != 5000 {
		t.Errorf("expected Month 24 to be 5000, got %d", got)
	}
}

func TestCost_ProjectedAmount_AnnualStepPercent(t *testing.T) {
	cost := Cost{
		Name:               "Software Licenses",
		BaseAmountPerMonth: 1000,
		Growth: GrowthStrategy{
			Type:       AnnualStepPercent,
			AnnualRate: 0.10, // 10% increase every year
		},
	}

	// Year 1 (Months 0-11) should be base amount
	if got := cost.ProjectedAmount(0); got != 1000 {
		t.Errorf("expected Month 0 to be 1000, got %d", got)
	}
	if got := cost.ProjectedAmount(11); got != 1000 {
		t.Errorf("expected Month 11 to be 1000, got %d", got)
	}

	// Year 2 (Months 12-23) should be 1000 * 1.10 = 1100
	if got := cost.ProjectedAmount(12); got != 1100 {
		t.Errorf("expected Month 12 to be 1100, got %d", got)
	}
	if got := cost.ProjectedAmount(23); got != 1100 {
		t.Errorf("expected Month 23 to be 1100, got %d", got)
	}

	// Year 3 (Months 24-35) should be 1100 * 1.10 = 1210
	if got := cost.ProjectedAmount(24); got != 1210 {
		t.Errorf("expected Month 24 to be 1210, got %d", got)
	}
}

func TestPlan_AddExpenditures(t *testing.T) {
	plan := newValidPlan(t)

	// Add an OpEx
	err := plan.AddOpEx("Salaries", 10000, GrowthStrategy{Type: FlatGrowth})
	if err != nil {
		t.Errorf("expected no error adding OpEx, got %v", err)
	}

	// Add a COGS
	err = plan.AddCOGS("Server Hosting", 2000, GrowthStrategy{Type: AnnualStepPercent, AnnualRate: 0.05})
	if err != nil {
		t.Errorf("expected no error adding COGS, got %v", err)
	}

	// Validation checks
	if err := plan.AddOpEx("", 100, GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddCOGS("Bad Cost", -10, GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := plan.AddOpEx("Bad Growth", 100, GrowthStrategy{Type: "FakeGrowth"}); !errors.Is(err, ErrInvalidGrowthType) {
		t.Errorf("expected ErrInvalidGrowthType, got %v", err)
	}
}

func TestPlan_AddRevenue_Validation(t *testing.T) {
	plan := newValidPlan(t)

	// Validation checks
	if err := plan.AddRevenue("", 100, GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddRevenue("Bad Sales", -10, GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := plan.AddRevenue("Bad Growth", 100, GrowthStrategy{Type: "FakeGrowth"}); !errors.Is(err, ErrInvalidGrowthType) {
		t.Errorf("expected ErrInvalidGrowthType, got %v", err)
	}
}

func TestPlan_OutOfBoundsRetrieval(t *testing.T) {
	plan := newValidPlan(t)

	// Setup a continuous revenue curve so we have data to check against
	_ = plan.AddRevenue("Sales", 5000, GrowthStrategy{Type: FlatGrowth})

	// FIX: Pass duration 12 to all MonthlyNetCashFlow checks
	// 1. Valid End Month (Index 11)
	if got := plan.MonthlyNetCashFlow(11, 12); got != 5000 {
		t.Errorf("expected Month 11 to return 5000, got %d", got)
	}

	// 2. Invalid Negative Month (Time Travel)
	if got := plan.MonthlyNetCashFlow(-1, 12); got != 0 {
		t.Errorf("expected negative month to return 0, got %d", got)
	}

	// 3. Invalid Future Month (Index 12 is the 13th month, out of bounds)
	if got := plan.MonthlyNetCashFlow(12, 12); got != 0 {
		t.Errorf("expected future month to return 0, got %d", got)
	}
}
