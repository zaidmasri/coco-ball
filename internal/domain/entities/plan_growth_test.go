package entities

import (
	"errors"
	"testing"

	"uuid"
)

func TestNewCost(t *testing.T) {
	tests := []struct {
		testName    string
		costName    string
		baseAmount  Money
		growth      GrowthStrategy
		expectedErr error
	}{
		{
			testName:    "Valid cost",
			costName:    "Rent",
			baseAmount:  mustUSD(1500),
			growth:      GrowthStrategy{Type: FlatGrowth},
			expectedErr: nil,
		},
		{
			testName:    "Fails on empty name",
			costName:    "   ",
			baseAmount:  mustUSD(1500),
			growth:      GrowthStrategy{Type: FlatGrowth},
			expectedErr: ErrInvalidName,
		},
		{
			testName:    "Fails on negative amount",
			costName:    "Rent",
			baseAmount:  mustUSD(-1),
			growth:      GrowthStrategy{Type: FlatGrowth},
			expectedErr: ErrNegativeAmount,
		},
		{
			testName:    "Fails on invalid growth type",
			costName:    "Rent",
			baseAmount:  mustUSD(1500),
			growth:      GrowthStrategy{Type: "FakeGrowth"},
			expectedErr: ErrInvalidGrowthType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cost, err := NewCost(tt.costName, tt.baseAmount, tt.growth)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr != nil {
				return
			}
			if cost.ID() == uuid.Nil() {
				t.Error("expected NewCost to generate a non-nil id")
			}
		})
	}
}

func TestCost_SetID(t *testing.T) {
	cost, err := NewCost("Rent", mustUSD(1500), GrowthStrategy{Type: FlatGrowth})
	if err != nil {
		t.Fatalf("failed to create cost: %v", err)
	}
	originalID := cost.ID()

	newID := uuid.NewV7()
	cost.SetID(newID)

	if cost.ID() == originalID {
		t.Error("expected SetID to change the cost's id")
	}
	if cost.ID() != newID {
		t.Errorf("expected id %s, got %s", newID, cost.ID())
	}
}

func TestCost_Setters(t *testing.T) {
	var cost Cost
	cost.SetName("  Rent  ")
	cost.SetBaseAmountPerMonth(mustUSD(1500))
	cost.SetGrowth(GrowthStrategy{Type: AnnualStepPercent, AnnualRate: 0.05})

	if got := cost.Name(); got != "Rent" {
		t.Errorf("expected SetName to trim whitespace and store %q, got %q", "Rent", got)
	}
	if got := cost.BaseAmountPerMonth(); !got.Equal(mustUSD(1500)) {
		t.Errorf("expected BaseAmountPerMonth %s, got %s", mustUSD(1500), got)
	}
	if got := cost.Growth(); got.Type != AnnualStepPercent || got.AnnualRate != 0.05 {
		t.Errorf("expected Growth {AnnualStepPercent 0.05}, got %+v", got)
	}

	// Setters are deliberately unvalidated - reconstructing an incomplete
	// draft (e.g. a fresh wizard row with no name yet) must not error.
	cost.SetName("")
	if got := cost.Name(); got != "" {
		t.Errorf("expected SetName(\"\") to be allowed and store empty, got %q", got)
	}
}

func TestPlan_AddOpEx_AssignsID(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddOpEx("Rent", mustUSD(1500), GrowthStrategy{Type: FlatGrowth}); err != nil {
		t.Fatalf("AddOpEx failed: %v", err)
	}

	opEx := plan.OpEx()
	if len(opEx) != 1 {
		t.Fatalf("expected 1 opex item, got %d", len(opEx))
	}
	if opEx[0].ID() == uuid.Nil() {
		t.Error("expected AddOpEx to assign a non-nil id via NewCost")
	}
}

func TestPlan_TotalExpenses(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddCOGS("Metal", mustUSD(2000), GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Software", mustUSD(500), GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddCOGS("Wood", mustUSD(300), GrowthStrategy{Type: FlatGrowth})

	expectedTotal := mustUSD(33600)

	if got := plan.TotalExpenses(12); !got.Equal(expectedTotal) {
		t.Errorf("expected total %s, got %s", expectedTotal, got)
	}
}

func TestPlan_MonthlyLedgerMath(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddRevenue("Software Subs", mustUSD(5000), GrowthStrategy{Type: FlatGrowth})

	_ = plan.AddOpEx("Rent", mustUSD(1500), GrowthStrategy{Type: FlatGrowth})
	_ = plan.AddOpEx("Marketing", mustUSD(1000), GrowthStrategy{Type: FlatGrowth})

	if got := plan.MonthlyNetCashFlow(0, 12); got.MinorUnits() != 2500 {
		t.Errorf("expected Month 0 net flow of 2500, got %s", got)
	}
	if got := plan.MonthlyNetCashFlow(11, 12); got.MinorUnits() != 2500 {
		t.Errorf("expected Month 11 net flow of 2500, got %s", got)
	}

	if got := plan.MonthlyNetCashFlow(12, 12); !got.IsZero() {
		t.Errorf("expected Month 12 net flow of 0 (out of bounds), got %s", got)
	}

	if got := plan.TotalExpenses(12); got.MinorUnits() != 30000 {
		t.Errorf("expected Total Expenses of 30000, got %s", got)
	}

	if got := plan.TotalRevenues(12); got.MinorUnits() != 60000 {
		t.Errorf("expected Total Revenues of 60000, got %s", got)
	}
}

func TestCost_ProjectedAmount_Flat(t *testing.T) {
	cost, err := NewCost("Rent", mustUSD(5000), GrowthStrategy{Type: FlatGrowth})
	if err != nil {
		t.Fatalf("unexpected error creating cost: %v", err)
	}

	if got := cost.ProjectedAmount(0); got.MinorUnits() != 5000 {
		t.Errorf("expected Month 0 to be 5000, got %s", got)
	}
	if got := cost.ProjectedAmount(24); got.MinorUnits() != 5000 {
		t.Errorf("expected Month 24 to be 5000, got %s", got)
	}
}

func TestCost_ProjectedAmount_AnnualStepPercent(t *testing.T) {
	cost, err := NewCost("Software Licenses", mustUSD(1000), GrowthStrategy{
		Type:       AnnualStepPercent,
		AnnualRate: 0.10, // 10% increase every year
	})
	if err != nil {
		t.Fatalf("unexpected error creating cost: %v", err)
	}

	// Year 1 (Months 0-11) should be base amount
	if got := cost.ProjectedAmount(0); got.MinorUnits() != 1000 {
		t.Errorf("expected Month 0 to be 1000, got %s", got)
	}
	if got := cost.ProjectedAmount(11); got.MinorUnits() != 1000 {
		t.Errorf("expected Month 11 to be 1000, got %s", got)
	}

	// Year 2 (Months 12-23) should be 1000 * 1.10 = 1100
	if got := cost.ProjectedAmount(12); got.MinorUnits() != 1100 {
		t.Errorf("expected Month 12 to be 1100, got %s", got)
	}
	if got := cost.ProjectedAmount(23); got.MinorUnits() != 1100 {
		t.Errorf("expected Month 23 to be 1100, got %s", got)
	}

	// Year 3 (Months 24-35) should be 1100 * 1.10 = 1210
	if got := cost.ProjectedAmount(24); got.MinorUnits() != 1210 {
		t.Errorf("expected Month 24 to be 1210, got %s", got)
	}
}

func TestPlan_AddExpenditures(t *testing.T) {
	plan := newValidPlan(t)

	err := plan.AddOpEx("Salaries", mustUSD(10000), GrowthStrategy{Type: FlatGrowth})
	if err != nil {
		t.Errorf("expected no error adding OpEx, got %v", err)
	}

	err = plan.AddCOGS("Server Hosting", mustUSD(2000), GrowthStrategy{Type: AnnualStepPercent, AnnualRate: 0.05})
	if err != nil {
		t.Errorf("expected no error adding COGS, got %v", err)
	}

	if err := plan.AddOpEx("", mustUSD(100), GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddCOGS("Bad Cost", mustUSD(-10), GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := plan.AddOpEx("Bad Growth", mustUSD(100), GrowthStrategy{Type: "FakeGrowth"}); !errors.Is(err, ErrInvalidGrowthType) {
		t.Errorf("expected ErrInvalidGrowthType, got %v", err)
	}
}

func TestPlan_AddRevenue_Validation(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddRevenue("", mustUSD(100), GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddRevenue("Bad Sales", mustUSD(-10), GrowthStrategy{Type: FlatGrowth}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := plan.AddRevenue("Bad Growth", mustUSD(100), GrowthStrategy{Type: "FakeGrowth"}); !errors.Is(err, ErrInvalidGrowthType) {
		t.Errorf("expected ErrInvalidGrowthType, got %v", err)
	}
}

func TestPlan_OutOfBoundsRetrieval(t *testing.T) {
	plan := newValidPlan(t)

	_ = plan.AddRevenue("Sales", mustUSD(5000), GrowthStrategy{Type: FlatGrowth})

	if got := plan.MonthlyNetCashFlow(11, 12); got.MinorUnits() != 5000 {
		t.Errorf("expected Month 11 to return 5000, got %s", got)
	}

	if got := plan.MonthlyNetCashFlow(-1, 12); !got.IsZero() {
		t.Errorf("expected negative month to return 0, got %s", got)
	}

	if got := plan.MonthlyNetCashFlow(12, 12); !got.IsZero() {
		t.Errorf("expected future month to return 0, got %s", got)
	}
}
