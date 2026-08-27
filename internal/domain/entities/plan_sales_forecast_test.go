package entities

import (
	"errors"
	"testing"
)

func TestValidateProduct(t *testing.T) {
	if err := ValidateProduct(Product{Name: "", PricePerUnit: mustUSD(10)}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", PricePerUnit: mustUSD(-1)}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", CostPerUnit: mustUSD(-1)}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", PricePerUnit: mustUSD(maxMoneyAmount + 1)}); !errors.Is(err, ErrAmountTooLarge) {
		t.Errorf("expected ErrAmountTooLarge, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", PricePerUnit: mustUSD(20), CostPerUnit: mustUSD(10)}); err != nil {
		t.Errorf("expected valid product, got %v", err)
	}
}

func TestValidateSalesGrowthCurve(t *testing.T) {
	tooLow := SalesGrowthCurve{Year1QuarterlyRates: [4]float64{-1.5, 0, 0, 0}}
	if err := ValidateSalesGrowthCurve(tooLow); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate for -150%% quarterly rate, got %v", err)
	}
	tooHigh := SalesGrowthCurve{FutureYearRates: []float64{11.0, 0}}
	if err := ValidateSalesGrowthCurve(tooHigh); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate for +1100%% future-year rate, got %v", err)
	}
	valid := SalesGrowthCurve{Year1QuarterlyRates: [4]float64{0.05, 0.05, 0.05, 0.05}, FutureYearRates: []float64{0.1, 0.1}}
	if err := ValidateSalesGrowthCurve(valid); err != nil {
		t.Errorf("expected valid sales growth curve, got %v", err)
	}
}

func TestPlan_AddProduct_Validation(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddProduct(Product{Name: "", PricePerUnit: mustUSD(10)}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddProduct(Product{Name: "Widget", PricePerUnit: mustUSD(20), CostPerUnit: mustUSD(10)}); err != nil {
		t.Errorf("expected no error adding product, got %v", err)
	}
	if got := len(plan.Products()); got != 1 {
		t.Errorf("expected 1 product, got %d", got)
	}
}
