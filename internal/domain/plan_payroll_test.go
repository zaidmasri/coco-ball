package domain

import (
	"errors"
	"testing"
)

func TestValidateSalaryRole(t *testing.T) {
	if err := ValidateSalaryRole(SalaryRole{Role: "", MonthlyPay: 5000}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateSalaryRole(SalaryRole{Role: "Owner", MonthlyPay: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateSalaryRole(SalaryRole{Role: "Owner", MonthlyPay: 5000}); err != nil {
		t.Errorf("expected valid salary role, got %v", err)
	}
}

func TestValidateBenefit(t *testing.T) {
	if err := ValidateBenefit(Benefit{Type: "", MonthlyAmount: 200}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateBenefit(Benefit{Type: "Health Insurance", MonthlyAmount: -1}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateBenefit(Benefit{Type: "Health Insurance", MonthlyAmount: 200}); err != nil {
		t.Errorf("expected valid benefit, got %v", err)
	}
}

func TestPlan_AddSalaryRole_Validation(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddSalaryRole(SalaryRole{Role: "", MonthlyPay: 100}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddSalaryRole(SalaryRole{Role: "Founder", MonthlyPay: 5000}); err != nil {
		t.Errorf("expected no error adding salary role, got %v", err)
	}
	if got := len(plan.SalaryRoles()); got != 1 {
		t.Errorf("expected 1 salary role, got %d", got)
	}
}
