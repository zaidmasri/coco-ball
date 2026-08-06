package domain

import (
	"errors"
	"testing"
)

func TestValidateProduct(t *testing.T) {
	if err := ValidateProduct(Product{Name: "", PricePerUnit: 10}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", PricePerUnit: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", CostPerUnit: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateProduct(Product{Name: "Widget", PricePerUnit: 20, CostPerUnit: 10}); err != nil {
		t.Errorf("expected valid product, got %v", err)
	}
}

func TestPlan_AddProduct_Validation(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddProduct(Product{Name: "", PricePerUnit: 10}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddProduct(Product{Name: "Widget", PricePerUnit: 20, CostPerUnit: 10}); err != nil {
		t.Errorf("expected no error adding product, got %v", err)
	}
	if got := len(plan.Products()); got != 1 {
		t.Errorf("expected 1 product, got %d", got)
	}
}
