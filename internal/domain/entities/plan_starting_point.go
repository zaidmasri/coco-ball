package entities

type FinancingTerm struct {
	PrincipalMoney Money
	InterestRate   float64 // e.g., 0.07 for 7%
	TermMonths     int
}

type CapitalAsset struct {
	Name               string
	PurchaseCost       Money
	UsefulLifeMonths   int
	SalvageValue       Money
	PurchaseMonthIndex MonthIndex
	DepreciationMethod DepreciationMethod
	AssociatedLoan     *FinancingTerm
}

type StartupCost struct {
	Name   string
	Amount Money
}

type FundingSource struct {
	Name         string
	Amount       Money
	InterestRate float64
	TermMonths   int
}

type StartingBalances struct {
	Cash               Money
	AccountsReceivable Money
	PrepaidExpenses    Money
	AccountsPayable    Money
	AccruedExpenses    Money
}

// DepreciationForMonth calculates the exact depreciation expense for a specific normalized month.
func (c CapitalAsset) DepreciationForMonth(month MonthIndex) Money {
	if month < c.PurchaseMonthIndex {
		return Money{}
	}

	if month >= c.PurchaseMonthIndex+MonthIndex(c.UsefulLifeMonths) {
		return Money{}
	}

	monthSincePurchase := int(month - c.PurchaseMonthIndex)

	if c.DepreciationMethod == StraightLine {
		depreciableBase := c.PurchaseCost.Sub(c.SalvageValue)
		if !depreciableBase.IsPositive() {
			return Money{}
		}
		return depreciableBase.Div(int64(c.UsefulLifeMonths))
	}

	if c.DepreciationMethod == DoubleDeclining {
		rate := 2.0 / float64(c.UsefulLifeMonths)
		bookValue := float64(c.PurchaseCost.MinorUnits())

		// Because DDB relies on the previous month's book value, we must iterate
		// up to the requested month to find the exact expense.
		for i := 0; i <= monthSincePurchase; i++ {
			expense := fromFloatUSD(bookValue * rate)

			// Edge Case: Salvage Value Floor
			if fromFloatUSD(bookValue).Sub(expense).Less(c.SalvageValue) {
				expense = fromFloatUSD(bookValue).Sub(c.SalvageValue)
			}

			if i == monthSincePurchase {
				return expense
			}

			bookValue -= float64(expense.MinorUnits())

			// Edge Case: Fully depreciated before useful life ends (common in DDB)
			if fromFloatUSD(bookValue).LessOrEqual(c.SalvageValue) {
				return Money{}
			}
		}
	}

	return Money{}
}

// LoadStartingPointData populates the Starting Point section fields from an
// external source. It exists for the store layer's use: since Starting
// Point moved to normalized SQL tables, Plan's own JSON (un)marshalling no
// longer carries this data, so the store queries the tables itself and
// hands the results back here after loading the rest of the Plan.
func (p *Plan) LoadStartingPointData(assets []CapitalAsset, costs []StartupCost, funding []FundingSource, balances StartingBalances) {
	p.futurePurchases = assets
	p.startupCosts = costs
	p.fundingSources = funding
	p.startingBalances = balances
}

// ValidateStartupCost checks a StartupCost before it's persisted.
func ValidateStartupCost(cost StartupCost) error {
	if _, err := validateRequiredName(cost.Name); err != nil {
		return err
	}
	if err := validateMoneyAmount(cost.Amount); err != nil {
		return err
	}
	return nil
}

// AddStartupCost validates and appends a startup cost line item.
func (p *Plan) AddStartupCost(name string, amount Money) error {
	cost := StartupCost{Name: name, Amount: amount}
	if err := ValidateStartupCost(cost); err != nil {
		return err
	}
	p.startupCosts = append(p.startupCosts, cost)
	return nil
}

// ValidateFundingSource checks a FundingSource before it's persisted.
func ValidateFundingSource(funding FundingSource) error {
	if _, err := validateRequiredName(funding.Name); err != nil {
		return err
	}
	if err := validateMoneyAmount(funding.Amount); err != nil {
		return err
	}
	if err := validatePercentRate(funding.InterestRate); err != nil {
		return err
	}
	if funding.TermMonths < 0 || funding.TermMonths > 600 {
		return ErrInvalidTerm
	}
	return nil
}

// AddFundingSource validates and appends a funding source line item.
func (p *Plan) AddFundingSource(name string, amount Money, rate float64, term int) error {
	funding := FundingSource{Name: name, Amount: amount, InterestRate: rate, TermMonths: term}
	if err := ValidateFundingSource(funding); err != nil {
		return err
	}
	p.fundingSources = append(p.fundingSources, funding)
	return nil
}

func (p *Plan) SetStartingBalances(cash, ar, pe, ap, ae Money) {
	p.startingBalances = StartingBalances{
		Cash:               cash,
		AccountsReceivable: ar,
		PrepaidExpenses:    pe,
		AccountsPayable:    ap,
		AccruedExpenses:    ae,
	}
}

// ValidateCapitalAsset checks a CapitalAsset before it's persisted.
func ValidateCapitalAsset(asset CapitalAsset) error {
	if _, err := validateRequiredName(asset.Name); err != nil {
		return err
	}
	if err := validateMoneyAmount(asset.PurchaseCost); err != nil {
		return err
	}
	if err := validateMoneyAmount(asset.SalvageValue); err != nil {
		return err
	}
	if asset.DepreciationMethod != None && (asset.UsefulLifeMonths < 1 || asset.UsefulLifeMonths > 600) {
		return ErrInvalidUsefulLife
	}
	if asset.PurchaseCost.Less(asset.SalvageValue) {
		return ErrPurchaseCostLessThanSalvageValue
	}
	if asset.DepreciationMethod != StraightLine && asset.DepreciationMethod != DoubleDeclining && asset.DepreciationMethod != None {
		return ErrInvalidDepreciationMethod
	}
	return nil
}

// ValidateStartingBalances checks the Cash on Hand balances before the
// singleton section is marked complete.
func ValidateStartingBalances(balances StartingBalances) error {
	for _, amt := range []Money{balances.Cash, balances.AccountsReceivable, balances.PrepaidExpenses, balances.AccountsPayable, balances.AccruedExpenses} {
		if err := validateMoneyAmount(amt); err != nil {
			return err
		}
	}
	return nil
}

// AddCapitalPurchase validates and appends a new asset.
func (p *Plan) AddCapitalPurchase(asset CapitalAsset) error {
	if err := ValidateCapitalAsset(asset); err != nil {
		return err
	}
	p.futurePurchases = append(p.futurePurchases, asset)
	return nil
}

func (p *Plan) StartupCosts() []StartupCost        { return p.startupCosts }
func (p *Plan) FundingSources() []FundingSource    { return p.fundingSources }
func (p *Plan) StartingBalances() StartingBalances { return p.startingBalances }

// UsefulLifeYears formatting Helpers for the HTML Templates ---
// The form takes Years, but the domain stores Months. This converts it back for the form.
func (c CapitalAsset) UsefulLifeYears() int {
	return c.UsefulLifeMonths / 12
}

// InterestRatePercent handles the form which takes 5.5%, but the domain stores 0.055. This converts it back for the form.
func (f FundingSource) InterestRatePercent() float64 {
	return f.InterestRate * 100
}

func (p *Plan) FuturePurchases() []CapitalAsset {
	res := make([]CapitalAsset, len(p.futurePurchases))
	copy(res, p.futurePurchases)
	return res
}
