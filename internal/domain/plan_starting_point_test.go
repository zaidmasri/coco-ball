package domain

import (
	"errors"
	"testing"
)

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

func TestValidateStartupCost(t *testing.T) {
	if err := ValidateStartupCost(StartupCost{Name: "", Amount: 100}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateStartupCost(StartupCost{Name: "Legal Fees", Amount: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateStartupCost(StartupCost{Name: "Legal Fees", Amount: 500}); err != nil {
		t.Errorf("expected valid startup cost, got %v", err)
	}
}

func TestValidateFundingSource(t *testing.T) {
	if err := ValidateFundingSource(FundingSource{Name: "", Amount: 100}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateFundingSource(FundingSource{Name: "SBA Loan", Amount: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateFundingSource(FundingSource{Name: "SBA Loan", Amount: 50000}); err != nil {
		t.Errorf("expected valid funding source, got %v", err)
	}
}
