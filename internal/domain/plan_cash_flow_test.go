package domain

import (
	"errors"
	"testing"
)

func TestValidateInventoryPurchase(t *testing.T) {
	if err := ValidateInventoryPurchase(InventoryPurchase{Category: "", MonthlyAmount: 100}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateInventoryPurchase(InventoryPurchase{Category: "Raw Materials", MonthlyAmount: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateInventoryPurchase(InventoryPurchase{Category: "Raw Materials", MonthlyAmount: 100}); err != nil {
		t.Errorf("expected valid inventory purchase, got %v", err)
	}
}

func TestValidateDistribution(t *testing.T) {
	if err := ValidateDistribution(Distribution{Name: "", MonthlyAmount: 100}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateDistribution(Distribution{Name: "Owner's Distribution", MonthlyAmount: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateDistribution(Distribution{Name: "Owner's Distribution", MonthlyAmount: 100}); err != nil {
		t.Errorf("expected valid distribution, got %v", err)
	}
}
