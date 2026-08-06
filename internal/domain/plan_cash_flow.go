package domain

import "strings"

// InventoryPurchase is a discretionary additional-inventory cash outflow.
type InventoryPurchase struct {
	Category       string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

// Distribution is a discretionary owner distribution / repayment cash outflow.
type Distribution struct {
	Name           string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (i InventoryPurchase) GrowthRatePercent(index int) float64 {
	return i.GrowthAfterYr1.GrowthRatePercent(index)
}

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (d Distribution) GrowthRatePercent(index int) float64 {
	return d.GrowthAfterYr1.GrowthRatePercent(index)
}

// LoadCashFlowData populates the discretionary Cash Flow section fields
// from an external source. Mirrors LoadStartingPointData: Cash Flow moved
// to normalized SQL tables, so Plan's own JSON (un)marshalling no longer
// carries this data.
func (p *Plan) LoadCashFlowData(inventory []InventoryPurchase, distributions []Distribution) {
	p.inventoryPlan = inventory
	p.distributions = distributions
}

// ClearCashFlow wipes existing discretionary cash flow data.
func (p *Plan) ClearCashFlow() {
	p.inventoryPlan = make([]InventoryPurchase, 0)
	p.distributions = make([]Distribution, 0)
}

// ValidateInventoryPurchase checks an InventoryPurchase before it's persisted.
func ValidateInventoryPurchase(inv InventoryPurchase) error {
	if strings.TrimSpace(inv.Category) == "" {
		return ErrInvalidName
	}
	if inv.MonthlyAmount < 0 {
		return ErrNegativeAmount
	}
	return nil
}

// AddInventoryPurchase appends an additional-inventory cash outflow line.
func (p *Plan) AddInventoryPurchase(inv InventoryPurchase) error {
	if err := ValidateInventoryPurchase(inv); err != nil {
		return err
	}
	p.inventoryPlan = append(p.inventoryPlan, inv)
	return nil
}

// ValidateDistribution checks a Distribution before it's persisted.
func ValidateDistribution(dist Distribution) error {
	if strings.TrimSpace(dist.Name) == "" {
		return ErrInvalidName
	}
	if dist.MonthlyAmount < 0 {
		return ErrNegativeAmount
	}
	return nil
}

// AddDistribution appends an owner distribution / repayment cash outflow line.
func (p *Plan) AddDistribution(dist Distribution) error {
	if err := ValidateDistribution(dist); err != nil {
		return err
	}
	p.distributions = append(p.distributions, dist)
	return nil
}

func (p *Plan) AdditionalInventory() []InventoryPurchase { return p.inventoryPlan }
func (p *Plan) Distributions() []Distribution            { return p.distributions }
