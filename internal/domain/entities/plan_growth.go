package entities

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

// AnnualGrowth captures a Year 2 / Year 3 (etc.) growth rate schedule that
// applies on top of a Year 1 base amount. Index 0 = Year 2, index 1 = Year 3.
type AnnualGrowth struct {
	RatesAfterYear1 []float64
}

// GrowthRatePercent returns the growth rate at the given index expressed as a
// percentage (e.g. 0.03 -> 3.0), for use directly in HTML templates.
func (g AnnualGrowth) GrowthRatePercent(index int) float64 {
	if index < 0 || index >= len(g.RatesAfterYear1) {
		return 0
	}
	return g.RatesAfterYear1[index] * 100
}

type GrowthStrategy struct {
	Type       GrowthType
	AnnualRate float64
}

// AnnualRatePercent returns the annual growth rate as a percentage (e.g. 0.03 -> 3.0).
func (g GrowthStrategy) AnnualRatePercent() float64 {
	return g.AnnualRate * 100
}

// Cost is a repeatable line item used for both Operating Expenses and COGS.
// It carries its own identity (id) so a wizard row and the domain value it
// stores can be referenced by the same UUID — see NewCost.
type Cost struct {
	id                 uuid.UUID
	Name               string
	BaseAmountPerMonth Money
	Growth             GrowthStrategy
}

func (c Cost) ID() uuid.UUID { return c.id }

// SetID overrides a Cost's identity. Used only by repository implementations
// reconstructing a Cost already persisted under a known ID (the wizard
// item's row ID) — mirrors User.SetID's reconstruction pattern. Reconstructed
// rows may be incomplete drafts, so this deliberately bypasses NewCost's
// validation.
func (c *Cost) SetID(id uuid.UUID) { c.id = id }

// validate checks a cost's business invariants (mirrors Plan.validate()).
func (c Cost) validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrInvalidName
	}
	if c.BaseAmountPerMonth.IsNegative() {
		return ErrNegativeAmount
	}
	if c.Growth.Type != FlatGrowth && c.Growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}
	return nil
}

// NewCost creates a new Cost line item with a domain-generated UUIDv7
// identity, validating name/amount/growth against the same invariants a
// persisted Cost must satisfy. Mirrors NewPlan's shape: generate the id,
// build the value, validate before returning — invalid states never leave
// this constructor. Use SetID (via a repository implementation) instead when
// reconstructing an already-persisted, possibly-incomplete draft.
func NewCost(name string, baseAmount Money, growth GrowthStrategy) (Cost, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Cost{}, fmt.Errorf("failed to generate cost id: %w", err)
	}

	c := Cost{
		id:                 id,
		Name:               strings.TrimSpace(name),
		BaseAmountPerMonth: baseAmount,
		Growth:             growth,
	}
	if err := c.validate(); err != nil {
		return Cost{}, err
	}
	return c, nil
}

func (c Cost) ProjectedAmount(month MonthIndex) Money {
	if c.Growth.Type == AnnualStepPercent && c.Growth.AnnualRate > 0 {
		yearsPassed := int(month) / 12
		if yearsPassed > 0 {
			multiplier := math.Pow(1.0+c.Growth.AnnualRate, float64(yearsPassed))
			return c.BaseAmountPerMonth.Mul(multiplier)
		}
	}
	return c.BaseAmountPerMonth
}

type RevenueStream struct {
	Name               string
	BaseAmountPerMonth Money
	Growth             GrowthStrategy
}

func (r RevenueStream) ProjectedAmount(month MonthIndex) Money {
	if r.Growth.Type == AnnualStepPercent && r.Growth.AnnualRate > 0 {
		yearsPassed := int(month) / 12
		if yearsPassed > 0 {
			multiplier := math.Pow(1.0+r.Growth.AnnualRate, float64(yearsPassed))
			return r.BaseAmountPerMonth.Mul(multiplier)
		}
	}
	return r.BaseAmountPerMonth
}

func (p *Plan) Revenues() []RevenueStream {
	res := make([]RevenueStream, len(p.revenues))
	copy(res, p.revenues)
	return res
}

// MonthlyRevenue calculates the projected revenue for a specific month
func (p *Plan) MonthlyRevenue(month MonthIndex) Money {
	var total Money
	for _, rev := range p.revenues {
		total = total.Add(rev.ProjectedAmount(month))
	}
	return total
}

// TotalRevenues now loops through the timeline to sum the continuous curves.
func (p *Plan) TotalRevenues(duration int) Money {
	var total Money
	for i := range duration {
		total = total.Add(p.MonthlyRevenue(MonthIndex(i)))
	}
	return total
}

// MonthlyDepreciation calculates the total non-cash depreciation expense for all assets in a given month.
func (p *Plan) MonthlyDepreciation(month MonthIndex) Money {
	var total Money
	for _, asset := range p.futurePurchases {
		total = total.Add(asset.DepreciationForMonth(month))
	}
	return total
}

func (p *Plan) MonthlyOpEx(month MonthIndex) Money {
	var total Money
	for _, exp := range p.opEx {
		total = total.Add(exp.ProjectedAmount(month))
	}
	return total
}

// MonthlyCOGS calculates the projected COGS for a specific month
func (p *Plan) MonthlyCOGS(month MonthIndex) Money {
	var total Money
	for _, cogs := range p.cogs {
		total = total.Add(cogs.ProjectedAmount(month))
	}
	return total
}

func (p *Plan) MonthlyNetCashFlow(month MonthIndex, duration int) Money {
	if int(month) < 0 || int(month) >= duration {
		return Money{}
	}
	return p.MonthlyRevenue(month).Sub(p.MonthlyOpEx(month)).Sub(p.MonthlyCOGS(month))
}

// TotalOpEx calculates the lifetime operating expenses over the plan's duration.
func (p *Plan) TotalOpEx(duration int) Money {
	var total Money
	for i := range duration {
		total = total.Add(p.MonthlyOpEx(MonthIndex(i)))
	}
	return total
}

// TotalCOGS calculates the lifetime COGS over the plan's duration.
func (p *Plan) TotalCOGS(duration int) Money {
	var total Money
	for i := range duration {
		total = total.Add(p.MonthlyCOGS(MonthIndex(i)))
	}
	return total
}

// TotalExpenses calculates the total lifetime expenses (OpEx + COGS).
func (p *Plan) TotalExpenses(duration int) Money {
	return p.TotalOpEx(duration).Add(p.TotalCOGS(duration))
}

func (p *Plan) AddRevenue(name string, baseAmount Money, growth GrowthStrategy) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if baseAmount.IsNegative() {
		return ErrNegativeAmount
	}
	if growth.Type != FlatGrowth && growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}

	p.revenues = append(p.revenues, RevenueStream{
		Name:               name,
		BaseAmountPerMonth: baseAmount,
		Growth:             growth,
	})
	return nil
}

// LoadOpExData populates the Operating Expenses field from an external
// source. Mirrors LoadCashFlowData: Operating Expenses moved to a
// normalized SQL table, so Plan's own JSON (un)marshalling no longer
// carries this data.
func (p *Plan) LoadOpExData(costs []Cost) {
	p.opEx = costs
}

// ClearOpEx wipes existing operating expense data so forms can cleanly overwrite it.
func (p *Plan) ClearOpEx() {
	p.opEx = make([]Cost, 0)
}

func (p *Plan) OpEx() []Cost {
	res := make([]Cost, len(p.opEx))
	copy(res, p.opEx)
	return res
}

func (p *Plan) AddOpEx(name string, baseAmount Money, growth GrowthStrategy) error {
	cost, err := NewCost(name, baseAmount, growth)
	if err != nil {
		return err
	}
	p.opEx = append(p.opEx, cost)
	return nil
}

func (p *Plan) AddCOGS(name string, baseAmount Money, growth GrowthStrategy) error {
	cost, err := NewCost(name, baseAmount, growth)
	if err != nil {
		return err
	}
	p.cogs = append(p.cogs, cost)
	return nil
}
