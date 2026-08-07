package domain

import "strings"

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
	// Edge Case: Asset hasn't been purchased yet
	if month < c.PurchaseMonthIndex {
		return 0
	}

	// Edge Case: Asset is past its useful life
	if month >= c.PurchaseMonthIndex+MonthIndex(c.UsefulLifeMonths) {
		return 0
	}

	monthSincePurchase := int(month - c.PurchaseMonthIndex)

	if c.DepreciationMethod == StraightLine {
		depreciableBase := c.PurchaseCost - c.SalvageValue
		if depreciableBase <= 0 {
			return 0
		}
		return depreciableBase / Money(c.UsefulLifeMonths)
	}

	if c.DepreciationMethod == DoubleDeclining {
		rate := 2.0 / float64(c.UsefulLifeMonths)
		bookValue := float64(c.PurchaseCost)

		// Because DDB relies on the previous month's book value, we must iterate
		// up to the requested month to find the exact expense.
		for i := 0; i <= monthSincePurchase; i++ {
			expense := Money(bookValue * rate)

			// Edge Case: Salvage Value Floor
			if Money(bookValue)-expense < c.SalvageValue {
				expense = Money(bookValue) - c.SalvageValue
			}

			// If we've reached the target month, return the calculated expense
			if i == monthSincePurchase {
				return expense
			}

			// Otherwise, reduce book value and continue to the next month
			bookValue -= float64(expense)

			// Edge Case: Fully depreciated before useful life ends (common in DDB)
			if Money(bookValue) <= c.SalvageValue {
				return 0
			}
		}
	}

	return 0
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
	if strings.TrimSpace(cost.Name) == "" {
		return ErrInvalidName
	}
	if cost.Amount < 0 {
		return ErrNegativeAmount
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
	if strings.TrimSpace(funding.Name) == "" {
		return ErrInvalidName
	}
	if funding.Amount < 0 {
		return ErrNegativeAmount
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

// ValidateCapitalAsset checks a CapitalAsset before it's persisted. Factored
// out of AddCapitalPurchase so the Starting Point wizard's step handlers
// can validate a single field/struct at a time without needing a *Plan
// receiver.
func ValidateCapitalAsset(asset CapitalAsset) error {
	if asset.PurchaseCost < 0 || asset.SalvageValue < 0 {
		return ErrNegativeAmount
	}
	if asset.DepreciationMethod != None && asset.UsefulLifeMonths < 1 {
		return ErrInvalidUsefulLife
	}
	if asset.PurchaseCost < asset.SalvageValue {
		return ErrPurchaseCostLessThanSalvageValue
	}
	if asset.DepreciationMethod != StraightLine && asset.DepreciationMethod != DoubleDeclining && asset.DepreciationMethod != None {
		return ErrInvalidDepreciationMethod
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
