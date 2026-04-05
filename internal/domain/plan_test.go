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

func TestPlan_TotalExpenses(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddCOGS("Metal", 2000, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Software", 500, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddCOGS("Wood", 300, GrowthStrategy{Type: FlatGrowth})

	// expectedTotal := Money(2800)
	expectedTotal := Money(33600)

	if got := plan.TotalExpenses(); got != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, got)
	}
}

func TestPlan_MonthIndexBounds(t *testing.T) {
	plan := newValidPlan(t)

	// Valid Month (Month 0 is the 1st month)
	err := plan.AddOpEx("Valid Rent", 1000, GrowthStrategy{Type: FlatGrowth})
	if err != nil {
		t.Errorf("expected no error for MonthIndex 0, got %v", err)
	}

	// Valid End Month (Month 11 is the 12th month)
	err = plan.AddRevenue("Valid Sales", 5000, 11)
	if err != nil {
		t.Errorf("expected no error for MonthIndex 11, got %v", err)
	}

	// Invalid Future Month
	err = plan.AddRevenue("Distant Future", 100, 12)
	if !errors.Is(err, ErrInvalidMonthIndex) {
		t.Errorf("expected ErrInvalidMonthIndex for 12, got %v", err)
	}
}

func TestPlan_MonthlyLedgerMath(t *testing.T) {
	plan := newValidPlan(t) // This is a 12-month plan

	// 1. Setup Continuous Costs (Apply to ALL months automatically)
	_ = plan.AddOpEx("Rent", 1500, GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Marketing", 1000, GrowthStrategy{Type: FlatGrowth})
	// Total OpEx per month = 2500

	// 2. Setup Discrete Revenues
	_ = plan.AddRevenue("Sales", 2000, 0) // Month 0 (Loss)
	_ = plan.AddRevenue("Sales", 5000, 1) // Month 1 (Profit)

	// --- Monthly Assertions ---

	// Month 0: Revenue (2000) - OpEx (2500) = -500
	if got := plan.MonthlyNetCashFlow(0); got != -500 {
		t.Errorf("expected Month 0 net flow of -500, got %d", got)
	}

	// Month 1: Revenue (5000) - OpEx (2500) = 2500
	if got := plan.MonthlyNetCashFlow(1); got != 2500 {
		t.Errorf("expected Month 1 net flow of 2500, got %d", got)
	}

	// Month 2: Revenue (0) - OpEx (2500) = -2500
	if got := plan.MonthlyNetCashFlow(2); got != -2500 {
		t.Errorf("expected Month 2 net flow of -2500, got %d", got)
	}

	// --- Lifetime Assertions ---

	// Total Expenses: 2500/mo * 12 months = 30000
	if got := plan.TotalExpenses(); got != 30000 {
		t.Errorf("expected Total Expenses of 30000, got %d", got)
	}

	// Total Revenues: 2000 + 5000 = 7000
	if got := plan.TotalRevenues(); got != 7000 {
		t.Errorf("expected Total Revenues of 7000, got %d", got)
	}
}

func TestCapitalAsset_StraightLine(t *testing.T) {
	asset := CapitalAsset{
		Name:               "Oven",
		PurchaseCost:       12000, // $12,000
		SalvageValue:       2000,  // $2,000
		UsefulLifeMonths:   10,    // (12000 - 2000) / 10 = 1000/mo
		PurchaseMonthIndex: 2,     // Bought in Month 2
		DepreciationMethod: StraightLine,
	}

	// Month 1: Hasn't been bought yet
	if got := asset.DepreciationForMonth(1); got != 0 {
		t.Errorf("expected 0 before purchase, got %d", got)
	}

	// Month 2: First month of depreciation
	if got := asset.DepreciationForMonth(2); got != 1000 {
		t.Errorf("expected 1000, got %d", got)
	}

	// Month 11: Last month of depreciation (Month 2 + 10 months = index 11 is the last valid month)
	if got := asset.DepreciationForMonth(11); got != 1000 {
		t.Errorf("expected 1000, got %d", got)
	}

	// Month 12: Fully depreciated
	if got := asset.DepreciationForMonth(12); got != 0 {
		t.Errorf("expected 0 after useful life, got %d", got)
	}
}

func TestCapitalAsset_DoubleDecliningBalance(t *testing.T) {
	asset := CapitalAsset{
		Name:               "Delivery Truck",
		PurchaseCost:       10000,
		SalvageValue:       6000,
		UsefulLifeMonths:   5, // Rate = 2 / 5 = 0.4 (40% per month)
		PurchaseMonthIndex: 0,
		DepreciationMethod: DoubleDeclining,
	}

	// Month 0: Book = 10000. Exp = 10000 * 0.4 = 4000. New Book = 6000.
	if got := asset.DepreciationForMonth(0); got != 4000 {
		t.Errorf("expected 4000 for Month 0, got %d", got)
	}

	// Month 1: Book = 6000. This is equal to the Salvage Value.
	// Therefore, depreciation MUST floor at 0.
	if got := asset.DepreciationForMonth(1); got != 0 {
		t.Errorf("expected 0 due to salvage floor, got %d", got)
	}
}

func TestPlan_AddCapitalPurchase_Validation(t *testing.T) {
	plan := newValidPlan(t)

	err := plan.AddCapitalPurchase(CapitalAsset{
		Name:               "Invalid Life",
		PurchaseCost:       1000,
		UsefulLifeMonths:   0, // Fails here
		DepreciationMethod: StraightLine,
	})
	if !errors.Is(err, ErrInvalidUsefulLife) {
		t.Errorf("expected ErrInvalidUsefulLife, got %v", err)
	}

	err = plan.AddCapitalPurchase(CapitalAsset{
		Name:               "Invalid Method",
		PurchaseCost:       1000,
		UsefulLifeMonths:   12,
		DepreciationMethod: "MadeUpMethod", // Fails here
	})
	if !errors.Is(err, ErrInvalidDepreciationMethod) {
		t.Errorf("expected ErrInvalidDepreciationMethod, got %v", err)
	}

	// Future Asset validation (Should Pass)
	err = plan.AddCapitalPurchase(CapitalAsset{
		Name:               "Future Asset",
		PurchaseCost:       1000,
		UsefulLifeMonths:   12,
		PurchaseMonthIndex: 999, // Way past plan duration
		DepreciationMethod: StraightLine,
	})
	if err != nil {
		t.Errorf("expected future asset to be valid, got %v", err)
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
