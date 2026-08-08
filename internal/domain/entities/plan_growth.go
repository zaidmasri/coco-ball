package entities

import (
	"math"
	"strings"
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

type Cost struct {
	Name               string
	BaseAmountPerMonth Money
	Growth             GrowthStrategy
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

// ValidateOpExpense checks a Cost before it's persisted as an Operating
// Expense line item.
func ValidateOpExpense(cost Cost) error {
	if strings.TrimSpace(cost.Name) == "" {
		return ErrInvalidName
	}
	if cost.BaseAmountPerMonth.IsNegative() {
		return ErrNegativeAmount
	}
	if cost.Growth.Type != FlatGrowth && cost.Growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}
	return nil
}

func (p *Plan) AddOpEx(name string, baseAmount Money, growth GrowthStrategy) error {
	cost := Cost{Name: name, BaseAmountPerMonth: baseAmount, Growth: growth}
	if err := ValidateOpExpense(cost); err != nil {
		return err
	}
	p.opEx = append(p.opEx, cost)
	return nil
}

func (p *Plan) AddCOGS(name string, baseAmount Money, growth GrowthStrategy) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if baseAmount.IsNegative() {
		return ErrNegativeAmount
	}
	if growth.Type != FlatGrowth && growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}

	p.cogs = append(p.cogs, Cost{Name: name, BaseAmountPerMonth: baseAmount, Growth: growth})
	return nil
}
