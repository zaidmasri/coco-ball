// Package domain contains the aggregate root for the business plan
package domain

import (
	"errors"
	"math"
	"strings"
)

type (
	// MonthIndex normalizes time. 0 = Month 1 of operations, 1 = Month 2, etc.
	MonthIndex         int64
	Money              int64
	DepreciationMethod string
	GrowthType         string
)

const (
	StraightLine    DepreciationMethod = "StraightLine"
	DoubleDeclining DepreciationMethod = "DoubleDeclining"

	FlatGrowth        GrowthType = "Flat"
	AnnualStepPercent GrowthType = "AnnualStepPercent"
)

var (
	ErrInvalidName                      = errors.New("name cannot be empty")
	ErrNegativeAmount                   = errors.New("expense amount cannot be negative")
	ErrInvalidMonthIndex                = errors.New("month index is out of bounds")
	ErrInvalidDuration                  = errors.New("plan duration must be at least 1 month")
	ErrInvalidDepreciationMethod        = errors.New("invalid deprecation method")
	ErrInvalidUsefulLife                = errors.New("invalid useful life")
	ErrPurchaseCostLessThanSalvageValue = errors.New("purchase cost cannot be less than salvage value")
	ErrInvalidGrowthType                = errors.New("invalid growth type")
)

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

type GrowthStrategy struct {
	Type       GrowthType
	AnnualRate float64
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
			// Compound the base amount annually: Base * (1 + rate)^years
			multiplier := math.Pow(1.0+c.Growth.AnnualRate, float64(yearsPassed))
			return Money(float64(c.BaseAmountPerMonth) * multiplier)
		}
	}

	return c.BaseAmountPerMonth
}

type Revenue struct {
	Name   string
	Amount Money
	Month  MonthIndex
}

type Plan struct {
	id              int
	name            string
	duration        int // Total months the projection spans
	revenues        []Revenue
	opEx            []Cost
	cogs            []Cost
	futurePurchases []CapitalAsset
}

func NewPlan(id int, name string, duration int) (*Plan, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}

	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	return &Plan{
		id:       id,
		name:     name,
		duration: duration,
		// Always initialize slices so they aren't nil
		revenues:        make([]Revenue, 0),
		opEx:            make([]Cost, 0),
		cogs:            make([]Cost, 0),
		futurePurchases: make([]CapitalAsset, 0),
	}, nil
}

func (p *Plan) ID() int       { return p.id }
func (p *Plan) Name() string  { return p.name }
func (p *Plan) Duration() int { return p.duration }

func (p *Plan) Revenues() []Revenue {
	res := make([]Revenue, len(p.revenues))
	copy(res, p.revenues)
	return res
}

func (p *Plan) FuturePurchases() []CapitalAsset {
	res := make([]CapitalAsset, len(p.futurePurchases))
	copy(res, p.futurePurchases)
	return res
}

func (p *Plan) MonthlyRevenue(month MonthIndex) Money {
	var total Money
	for _, rev := range p.revenues {
		if rev.Month == month {
			total += rev.Amount
		}
	}
	return total
}

// MonthlyDepreciation calculates the total non-cash depreciation expense for all assets in a given month.
func (p *Plan) MonthlyDepreciation(month MonthIndex) Money {
	var total Money
	for _, asset := range p.futurePurchases {
		total += asset.DepreciationForMonth(month)
	}
	return total
}

func (p *Plan) MonthlyOpEx(month MonthIndex) Money {
	var total Money
	for _, exp := range p.opEx {
		total += exp.ProjectedAmount(month)
	}
	return total
}

// MonthlyCOGS calculates the projected COGS for a specific month
func (p *Plan) MonthlyCOGS(month MonthIndex) Money {
	var total Money
	for _, cogs := range p.cogs {
		total += cogs.ProjectedAmount(month)
	}
	return total
}

func (p *Plan) MonthlyNetCashFlow(month MonthIndex) Money {
	if int(month) < 0 || int(month) >= p.duration {
		return 0
	}
	return p.MonthlyRevenue(month) - p.MonthlyOpEx(month) - p.MonthlyCOGS(month)
}

func (p *Plan) TotalRevenues() Money {
	var total Money = 0
	for _, exp := range p.revenues {
		total += exp.Amount
	}
	return total
}

// TotalOpEx calculates the lifetime operating expenses over the plan's duration.
func (p *Plan) TotalOpEx() Money {
	var total Money
	for i := 0; i < p.duration; i++ {
		total += p.MonthlyOpEx(MonthIndex(i))
	}
	return total
}

// TotalCOGS calculates the lifetime COGS over the plan's duration.
func (p *Plan) TotalCOGS() Money {
	var total Money
	for i := 0; i < p.duration; i++ {
		total += p.MonthlyCOGS(MonthIndex(i))
	}
	return total
}

// TotalExpenses calculates the total lifetime expenses (OpEx + COGS).
func (p *Plan) TotalExpenses() Money {
	return p.TotalOpEx() + p.TotalCOGS()
}

func (p *Plan) AddRevenue(name string, amount Money, month MonthIndex) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if amount < 0 {
		return ErrNegativeAmount
	}
	if int(month) < 0 || int(month) >= p.duration {
		return ErrInvalidMonthIndex
	}

	p.revenues = append(p.revenues, Revenue{Name: name, Amount: amount, Month: month})
	return nil
}

func (p *Plan) AddOpEx(name string, baseAmount Money, growth GrowthStrategy) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if baseAmount < 0 {
		return ErrNegativeAmount
	}
	if growth.Type != FlatGrowth && growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}

	p.opEx = append(p.opEx, Cost{Name: name, BaseAmountPerMonth: baseAmount, Growth: growth})
	return nil
}

func (p *Plan) AddCOGS(name string, baseAmount Money, growth GrowthStrategy) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if baseAmount < 0 {
		return ErrNegativeAmount
	}
	if growth.Type != FlatGrowth && growth.Type != AnnualStepPercent {
		return ErrInvalidGrowthType
	}

	p.cogs = append(p.cogs, Cost{Name: name, BaseAmountPerMonth: baseAmount, Growth: growth})
	return nil
}

// AddCapitalPurchase validates and appends a new asset.
func (p *Plan) AddCapitalPurchase(asset CapitalAsset) error {
	if asset.PurchaseCost < 0 || asset.SalvageValue < 0 {
		return ErrNegativeAmount
	}
	if asset.UsefulLifeMonths < 1 {
		return ErrInvalidUsefulLife
	}
	if asset.PurchaseCost < asset.SalvageValue {
		return ErrPurchaseCostLessThanSalvageValue
	}
	if asset.DepreciationMethod != StraightLine && asset.DepreciationMethod != DoubleDeclining {
		return ErrInvalidDepreciationMethod
	}

	// Note: We do NOT check if PurchaseMonthIndex > p.duration.
	// As requested, future assets outside the current projection are valid.

	p.futurePurchases = append(p.futurePurchases, asset)
	return nil
}
