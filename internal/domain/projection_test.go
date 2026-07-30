package domain

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

func newProjectionPlan(t *testing.T) *Plan {
	t.Helper()
	plan, err := NewPlan(uuid.New(), "Projection Co", 1, 2024, uuid.New())
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	return plan
}

func TestMonthlyProductRevenueAndCOGS_NoGrowth(t *testing.T) {
	plan := newProjectionPlan(t)
	if err := plan.AddProduct(Product{Name: "Widget", Month1Units: 100, PricePerUnit: 25, CostPerUnit: 10}); err != nil {
		t.Fatalf("AddProduct failed: %v", err)
	}

	months := plan.ProjectMonths(3)
	for i, mm := range months {
		if mm.Revenue != 2500 {
			t.Errorf("month %d: expected revenue 2500, got %d", i, mm.Revenue)
		}
		if mm.COGS != 1000 {
			t.Errorf("month %d: expected COGS 1000, got %d", i, mm.COGS)
		}
		if mm.GrossProfit != 1500 {
			t.Errorf("month %d: expected gross profit 1500, got %d", i, mm.GrossProfit)
		}
	}
}

func TestUnitGrowthMultipliers_QuarterlyAndFutureYears(t *testing.T) {
	plan := newProjectionPlan(t)
	plan.SetSalesGrowth(SalesGrowthCurve{
		Year1QuarterlyRates: [4]float64{0.10, 0.06, 0.04, 0.03},
		FutureYearRates:     []float64{0.03, 0.02},
	})
	if err := plan.AddProduct(Product{Name: "Widget", Month1Units: 100, PricePerUnit: 1, CostPerUnit: 0}); err != nil {
		t.Fatalf("AddProduct failed: %v", err)
	}

	months := plan.ProjectMonths(36)

	// Month 0 (Month 1): base, no growth applied yet.
	if got := months[0].Revenue; got != 100 {
		t.Errorf("month 0: expected revenue 100, got %d", got)
	}

	// Month 1 (Month 2, still Q1): 100 * 1.10 = 110
	if got := months[1].Revenue; got != 110 {
		t.Errorf("month 1: expected revenue 110, got %d", got)
	}

	// Month 2 (Month 3, still Q1): 110 * 1.10 = 121
	if got := months[2].Revenue; got != 121 {
		t.Errorf("month 2: expected revenue 121, got %d", got)
	}

	// Month 3 (Month 4, Q2 starts): 121 * 1.06 = 128.26 -> truncated to 128
	if got := months[3].Revenue; got != 128 {
		t.Errorf("month 3: expected revenue 128, got %d", got)
	}

	// Month 12 (start of Year 2): should apply FutureYearRates[0] (3%) relative to month 11.
	year1EndUnits := months[11].Revenue
	expectedMonth12 := Money(float64(year1EndUnits) * 1.03)
	if got := months[12].Revenue; got != expectedMonth12 {
		t.Errorf("month 12: expected revenue %d, got %d", expectedMonth12, got)
	}
}

func TestMonthlyPayrollCost_GrowthAndContractorTaxExemption(t *testing.T) {
	plan := newProjectionPlan(t)

	if err := plan.AddSalaryRole(SalaryRole{
		Role:           "Founder",
		IsContractor:   false,
		Headcount:      1,
		MonthlyPay:     5000,
		GrowthAfterYr1: AnnualGrowth{RatesAfterYear1: []float64{0.10, 0.05}},
	}); err != nil {
		t.Fatalf("AddSalaryRole failed: %v", err)
	}

	if err := plan.AddSalaryRole(SalaryRole{
		Role:         "Freelancer",
		IsContractor: true,
		Headcount:    1,
		MonthlyPay:   2000,
	}); err != nil {
		t.Fatalf("AddSalaryRole failed: %v", err)
	}

	if err := plan.AddBenefit(Benefit{Type: "Health", MonthlyAmount: 500}); err != nil {
		t.Fatalf("AddBenefit failed: %v", err)
	}

	plan.SetPayrollTaxRates(PayrollTaxRates{
		SocialSecurityRate: 0.062,
		MedicareRate:       0.0145,
	})

	months := plan.ProjectMonths(36)

	// Year 1: payroll cost = 5000 (founder) + 2000 (contractor) + 500 (benefits) = 7500
	if got := months[0].PayrollCost; got != 7500 {
		t.Errorf("month 0: expected payroll cost 7500, got %d", got)
	}
	// Payroll tax only on the W-2 founder salary: 5000 * (0.062+0.0145) = 382.5 -> truncated 382
	if got := months[0].PayrollTax; got != 382 {
		t.Errorf("month 0: expected payroll tax 382, got %d", got)
	}

	// Year 2 (month 12): founder salary grows 10% -> 5500; contractor and benefits unchanged.
	if got := months[12].PayrollCost; got != 5500+2000+500 {
		t.Errorf("month 12: expected payroll cost %d, got %d", 5500+2000+500, got)
	}
}

func TestAmortizationSchedule_PayoffAtTerm(t *testing.T) {
	schedule := amortizationSchedule(10000, 0.06, 12, 36)

	// Loan should be fully paid off at the end of the term.
	if got := schedule[11].EndingBalance; got != 0 {
		t.Errorf("expected loan paid off by month 11, ending balance %d", got)
	}

	// Months past the term should show no further activity.
	if schedule[12].Interest != 0 || schedule[12].Principal != 0 {
		t.Errorf("expected no activity after payoff, got interest=%d principal=%d", schedule[12].Interest, schedule[12].Principal)
	}

	// Sum of all principal payments should equal (approximately) the original principal.
	var totalPrincipal Money
	for _, e := range schedule {
		totalPrincipal += e.Principal
	}
	if diff := math.Abs(float64(totalPrincipal - 10000)); diff > 15 {
		t.Errorf("expected total principal ~10000 (within rounding), got %d", totalPrincipal)
	}
}

func TestBalanceSheetSnapshots_Balance(t *testing.T) {
	plan := newProjectionPlan(t)

	// Funding: one loan, one equity contribution.
	plan.AddFundingSource("Bank Loan", 50000, 0.07, 60)
	plan.AddFundingSource("Owner Investment", 20000, 0, 0)

	plan.AddStartupCost("Legal Fees", 2000)

	if err := plan.AddCapitalPurchase(CapitalAsset{
		Name:               "Equipment",
		PurchaseCost:       12000,
		UsefulLifeMonths:   60,
		DepreciationMethod: StraightLine,
	}); err != nil {
		t.Fatalf("AddCapitalPurchase failed: %v", err)
	}

	if err := plan.AddProduct(Product{Name: "Widget", Month1Units: 200, PricePerUnit: 20, CostPerUnit: 8}); err != nil {
		t.Fatalf("AddProduct failed: %v", err)
	}
	plan.SetSalesGrowth(SalesGrowthCurve{
		Year1QuarterlyRates: [4]float64{0.05, 0.03, 0.02, 0.01},
		FutureYearRates:     []float64{0.02, 0.01},
	})

	if err := plan.AddSalaryRole(SalaryRole{Role: "Founder", Headcount: 1, MonthlyPay: 4000}); err != nil {
		t.Fatalf("AddSalaryRole failed: %v", err)
	}
	plan.SetPayrollTaxRates(PayrollTaxRates{SocialSecurityRate: 0.062, MedicareRate: 0.0145})

	if err := plan.AddOpEx("Rent", 1500, GrowthStrategy{Type: FlatGrowth}); err != nil {
		t.Fatalf("AddOpEx failed: %v", err)
	}

	if err := plan.AddInventoryPurchase(InventoryPurchase{Category: "Raw Materials", MonthlyAmount: 300}); err != nil {
		t.Fatalf("AddInventoryPurchase failed: %v", err)
	}
	if err := plan.AddDistribution(Distribution{Name: "Owner Draw", MonthlyAmount: 500}); err != nil {
		t.Fatalf("AddDistribution failed: %v", err)
	}

	plan.SetStartingBalances(5000, 1000, 200, 800, 0)

	snapshots := plan.BalanceSheetSnapshots(3)
	for _, s := range snapshots {
		// Money is a whole-dollar int64, so dozens of monthly float64
		// truncations can drift by a few dollars over a multi-year
		// projection; anything beyond that indicates a real bug.
		diff := math.Abs(float64(s.TotalAssets - s.TotalLiabilitiesAndEquity))
		if diff > 25 {
			t.Errorf("year %d: balance sheet does not balance: assets=%d liabilities+equity=%d (diff %.0f)",
				s.Year, s.TotalAssets, s.TotalLiabilitiesAndEquity, diff)
		}
	}

	// Loan balance should be strictly decreasing year over year until payoff.
	if snapshots[1].LoansPayable >= snapshots[0].LoansPayable {
		t.Errorf("expected loan balance to decline from year 1 (%d) to year 2 (%d)", snapshots[0].LoansPayable, snapshots[1].LoansPayable)
	}

	// Owners equity should reflect only the non-loan funding source.
	if snapshots[0].OwnersEquity != 20000 {
		t.Errorf("expected owners equity 20000, got %d", snapshots[0].OwnersEquity)
	}
}

func TestBreakeven_ZeroRevenueGuardsAgainstDivideByZero(t *testing.T) {
	plan := newProjectionPlan(t)
	if err := plan.AddOpEx("Rent", 1000, GrowthStrategy{Type: FlatGrowth}); err != nil {
		t.Fatalf("AddOpEx failed: %v", err)
	}

	b := plan.Breakeven()
	if b.BreakevenAnnual != 0 {
		t.Errorf("expected breakeven to be 0 when there's no gross margin to fund it, got %d", b.BreakevenAnnual)
	}
}

func TestFinancialRatiosSeries_SalesGrowthYear1IsZero(t *testing.T) {
	plan := newProjectionPlan(t)
	if err := plan.AddProduct(Product{Name: "Widget", Month1Units: 100, PricePerUnit: 10, CostPerUnit: 4}); err != nil {
		t.Fatalf("AddProduct failed: %v", err)
	}

	ratios := plan.FinancialRatiosSeries(3)
	if ratios[0].SalesGrowthPercent != 0 {
		t.Errorf("expected Year 1 sales growth to be 0 (no prior year), got %f", ratios[0].SalesGrowthPercent)
	}
}
