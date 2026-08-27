package entities

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateRequiredName(t *testing.T) {
	if _, err := validateRequiredName("  "); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName for blank name, got %v", err)
	}
	if _, err := validateRequiredName(strings.Repeat("a", maxNameLength+1)); !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong for over-length name, got %v", err)
	}
	cleaned, err := validateRequiredName("  Widget  ")
	if err != nil {
		t.Fatalf("expected valid name, got %v", err)
	}
	if cleaned != "Widget" {
		t.Errorf("expected trimmed name %q, got %q", "Widget", cleaned)
	}
	if _, err := validateRequiredName(strings.Repeat("a", maxNameLength)); err != nil {
		t.Errorf("expected name at the exact max length to be valid, got %v", err)
	}
}

func TestValidateMoneyAmount(t *testing.T) {
	if err := validateMoneyAmount(mustUSD(-1)); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := validateMoneyAmount(mustUSD(maxMoneyAmount + 1)); !errors.Is(err, ErrAmountTooLarge) {
		t.Errorf("expected ErrAmountTooLarge, got %v", err)
	}
	if err := validateMoneyAmount(mustUSD(maxMoneyAmount)); err != nil {
		t.Errorf("expected amount at the exact max to be valid, got %v", err)
	}
	if err := validateMoneyAmount(mustUSD(0)); err != nil {
		t.Errorf("expected zero amount to be valid, got %v", err)
	}
}

func TestValidateGrowthRate(t *testing.T) {
	if err := validateGrowthRate(-1.01); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate below -100%%, got %v", err)
	}
	if err := validateGrowthRate(10.01); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate above +1000%%, got %v", err)
	}
	for _, rate := range []float64{-1.0, -0.5, 0, 0.05, 10.0} {
		if err := validateGrowthRate(rate); err != nil {
			t.Errorf("expected rate %v to be valid, got %v", rate, err)
		}
	}
}

func TestValidatePercentRate(t *testing.T) {
	if err := validatePercentRate(-0.01); !errors.Is(err, ErrInvalidRate) {
		t.Errorf("expected ErrInvalidRate for negative rate, got %v", err)
	}
	if err := validatePercentRate(1.01); !errors.Is(err, ErrInvalidRate) {
		t.Errorf("expected ErrInvalidRate above 100%%, got %v", err)
	}
	for _, rate := range []float64{0, 0.0765, 1.0} {
		if err := validatePercentRate(rate); err != nil {
			t.Errorf("expected rate %v to be valid, got %v", rate, err)
		}
	}
}

func TestValidateEmailFormat(t *testing.T) {
	invalid := []string{"not-an-email", "missing-domain@", "@missing-local.com", "no-at-sign.com"}
	for _, email := range invalid {
		if err := validateEmailFormat(email); !errors.Is(err, ErrInvalidEmailFormat) {
			t.Errorf("expected ErrInvalidEmailFormat for %q, got %v", email, err)
		}
	}
	if err := validateEmailFormat("user@example.com"); err != nil {
		t.Errorf("expected valid email to pass, got %v", err)
	}
}
