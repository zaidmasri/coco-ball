package entities

// SalaryRole is a single payroll line item (a role/title with headcount).
type SalaryRole struct {
	Role           string
	IsContractor   bool
	Headcount      int
	MonthlyPay     Money
	GrowthAfterYr1 AnnualGrowth
}

// Benefit is a single employee benefit line item.
type Benefit struct {
	Type           string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

// PayrollTaxRates holds employer-side payroll tax rates.
type PayrollTaxRates struct {
	SocialSecurityRate float64
	MedicareRate       float64
	FUTARate           float64
	SUTARate           float64
}

// LoadPayrollData populates the Payroll section fields from an external
// source. Mirrors LoadStartingPointData: Payroll moved to normalized SQL
// tables, so Plan's own JSON (un)marshalling no longer carries this data.
func (p *Plan) LoadPayrollData(roles []SalaryRole, benefits []Benefit, rates PayrollTaxRates) {
	p.salaryRoles = roles
	p.benefits = benefits
	p.payrollTaxRates = rates
}

// ClearPayroll wipes existing payroll data so forms can cleanly overwrite it.
func (p *Plan) ClearPayroll() {
	p.salaryRoles = make([]SalaryRole, 0)
	p.benefits = make([]Benefit, 0)
	p.payrollTaxRates = PayrollTaxRates{}
}

// ValidateSalaryRole checks a SalaryRole before it's persisted.
func ValidateSalaryRole(role SalaryRole) error {
	if _, err := validateRequiredName(role.Role); err != nil {
		return err
	}
	if err := validateMoneyAmount(role.MonthlyPay); err != nil {
		return err
	}
	for _, rate := range role.GrowthAfterYr1.RatesAfterYear1 {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// AddSalaryRole appends a salary/wage line item.
func (p *Plan) AddSalaryRole(role SalaryRole) error {
	if err := ValidateSalaryRole(role); err != nil {
		return err
	}
	p.salaryRoles = append(p.salaryRoles, role)
	return nil
}

// ValidateBenefit checks a Benefit before it's persisted.
func ValidateBenefit(benefit Benefit) error {
	if _, err := validateRequiredName(benefit.Type); err != nil {
		return err
	}
	if err := validateMoneyAmount(benefit.MonthlyAmount); err != nil {
		return err
	}
	for _, rate := range benefit.GrowthAfterYr1.RatesAfterYear1 {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePayrollTaxRates checks employer payroll tax rates before the
// singleton Payroll Tax Rates section is marked complete.
func ValidatePayrollTaxRates(rates PayrollTaxRates) error {
	for _, rate := range []float64{rates.SocialSecurityRate, rates.MedicareRate, rates.FUTARate, rates.SUTARate} {
		if err := validatePercentRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// AddBenefit appends an employee benefit line item.
func (p *Plan) AddBenefit(benefit Benefit) error {
	if err := ValidateBenefit(benefit); err != nil {
		return err
	}
	p.benefits = append(p.benefits, benefit)
	return nil
}

// SetPayrollTaxRates sets the employer payroll tax rates.
func (p *Plan) SetPayrollTaxRates(rates PayrollTaxRates) {
	p.payrollTaxRates = rates
}

func (p *Plan) SalaryRoles() []SalaryRole { return p.salaryRoles }
func (p *Plan) Benefits() []Benefit       { return p.benefits }
func (p *Plan) PayrollTaxRates() PayrollTaxRates {
	return p.payrollTaxRates
}

// HasPayrollTaxRates reports whether any payroll tax rate has been set,
// used by templates to decide whether to show suggested defaults.
func (r PayrollTaxRates) HasPayrollTaxRates() bool {
	return r.SocialSecurityRate != 0 || r.MedicareRate != 0 || r.FUTARate != 0 || r.SUTARate != 0
}

func (r PayrollTaxRates) SocialSecurityRatePercent() float64 { return r.SocialSecurityRate * 100 }
func (r PayrollTaxRates) MedicareRatePercent() float64       { return r.MedicareRate * 100 }
func (r PayrollTaxRates) FUTARatePercent() float64           { return r.FUTARate * 100 }
func (r PayrollTaxRates) SUTARatePercent() float64           { return r.SUTARate * 100 }

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (s SalaryRole) GrowthRatePercent(index int) float64 {
	return s.GrowthAfterYr1.GrowthRatePercent(index)
}

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (b Benefit) GrowthRatePercent(index int) float64 {
	return b.GrowthAfterYr1.GrowthRatePercent(index)
}
