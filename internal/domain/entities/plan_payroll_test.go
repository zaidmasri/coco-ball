package entities

import (
	"errors"
	"testing"
)

func TestValidateSalaryRole(t *testing.T) {
	if err := ValidateSalaryRole(SalaryRole{Role: "", MonthlyPay: mustUSD(5000)}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateSalaryRole(SalaryRole{Role: "Owner", MonthlyPay: mustUSD(-1)}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateSalaryRole(SalaryRole{Role: "Owner", MonthlyPay: mustUSD(5000), GrowthAfterYr1: AnnualGrowth{RatesAfterYear1: []float64{-5.0, 0}}}); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate for -500%% growth, got %v", err)
	}
	if err := ValidateSalaryRole(SalaryRole{Role: "Owner", MonthlyPay: mustUSD(5000)}); err != nil {
		t.Errorf("expected valid salary role, got %v", err)
	}
}

func TestValidateBenefit(t *testing.T) {
	if err := ValidateBenefit(Benefit{Type: "", MonthlyAmount: mustUSD(200)}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := ValidateBenefit(Benefit{Type: "Health Insurance", MonthlyAmount: mustUSD(-1)}); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("expected ErrNegativeAmount, got %v", err)
	}
	if err := ValidateBenefit(Benefit{Type: "Health Insurance", MonthlyAmount: mustUSD(200), GrowthAfterYr1: AnnualGrowth{RatesAfterYear1: []float64{11.0, 0}}}); !errors.Is(err, ErrInvalidGrowthRate) {
		t.Errorf("expected ErrInvalidGrowthRate for +1100%% growth, got %v", err)
	}
	if err := ValidateBenefit(Benefit{Type: "Health Insurance", MonthlyAmount: mustUSD(200)}); err != nil {
		t.Errorf("expected valid benefit, got %v", err)
	}
}

func TestValidatePayrollTaxRates(t *testing.T) {
	if err := ValidatePayrollTaxRates(PayrollTaxRates{SocialSecurityRate: -0.01}); !errors.Is(err, ErrInvalidRate) {
		t.Errorf("expected ErrInvalidRate for negative rate, got %v", err)
	}
	if err := ValidatePayrollTaxRates(PayrollTaxRates{MedicareRate: 1.5}); !errors.Is(err, ErrInvalidRate) {
		t.Errorf("expected ErrInvalidRate for rate over 100%%, got %v", err)
	}
	valid := PayrollTaxRates{SocialSecurityRate: 0.062, MedicareRate: 0.0145, FUTARate: 0.006, SUTARate: 0.034}
	if err := ValidatePayrollTaxRates(valid); err != nil {
		t.Errorf("expected valid payroll tax rates, got %v", err)
	}
}

func TestPlan_AddSalaryRole_Validation(t *testing.T) {
	plan := newValidPlan(t)

	if err := plan.AddSalaryRole(SalaryRole{Role: "", MonthlyPay: mustUSD(100)}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if err := plan.AddSalaryRole(SalaryRole{Role: "Founder", MonthlyPay: mustUSD(5000)}); err != nil {
		t.Errorf("expected no error adding salary role, got %v", err)
	}
	if got := len(plan.SalaryRoles()); got != 1 {
		t.Errorf("expected 1 salary role, got %d", got)
	}
}
