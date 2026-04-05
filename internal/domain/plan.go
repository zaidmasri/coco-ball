// Package domain contains the aggregate root for the business plan
package domain

import (
	"errors"
	"strings"
)

// MonthIndex normalizes time. 0 = Month 1 of operations, 1 = Month 2, etc.
type MonthIndex int64

type Money int64

type DepreciationMethod string

const (
	StraightLine    DepreciationMethod = "StraightLine"
	DoubleDeclining DepreciationMethod = "DoubleDeclining"
)

var (
	ErrInvalidName                      = errors.New("name cannot be empty")
	ErrNegativeAmount                   = errors.New("expense amount cannot be negative")
	ErrInvalidMonthIndex                = errors.New("month index is out of bounds")
	ErrInvalidDuration                  = errors.New("plan duration must be at least 1 month")
	ErrInvalidDepreciationMethod        = errors.New("invalid deprecation method")
	ErrInvalidUsefulLife                = errors.New("invalid useful life")
	ErrPurchaseCostLessThanSalvageValue = errors.New("purchase cost cannot be less than salvage value")
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
	// AssociatedLoan *FinancingTerm
}

// DepreciationForMonth calculates the exact depreciation expense for a specific normalized month.
func (c CapitalAsset) DepreciationForMonth(month MonthIndex) Money {
	// 1. Edge Case: Asset hasn't been purchased yet
	if month < c.PurchaseMonthIndex {
		return 0
	}

	// 2. Edge Case: Asset is past its useful life
	if month >= c.PurchaseMonthIndex+MonthIndex(c.UsefulLifeMonths) {
		return 0
	}

	monthSincePurchase := int(month - c.PurchaseMonthIndex)

	// --- Straight-Line Math ---
	if c.DepreciationMethod == StraightLine {
		depreciableBase := c.PurchaseCost - c.SalvageValue
		if depreciableBase <= 0 {
			return 0
		}
		return depreciableBase / Money(c.UsefulLifeMonths)
	}

	// --- Double Declining Balance Math ---
	if c.DepreciationMethod == DoubleDeclining {
		// DDB Rate = 2 / Useful Life
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

type Expense struct {
	Name   string
	Amount Money
	Month  MonthIndex
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
	expenses        []Expense
	revenues        []Revenue
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
		expenses:        make([]Expense, 0),
		revenues:        make([]Revenue, 0),
		futurePurchases: make([]CapitalAsset, 0),
	}, nil
}

func (p *Plan) AddExpense(name string, amount Money, month MonthIndex) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if amount < 0 {
		return ErrNegativeAmount
	}
	if int(month) < 0 || int(month) >= p.duration {
		return ErrInvalidMonthIndex
	}

	p.expenses = append(p.expenses, Expense{
		Name:   name,
		Amount: amount,
		Month:  month,
	})
	return nil
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

func (p *Plan) ID() int       { return p.id }
func (p *Plan) Name() string  { return p.name }
func (p *Plan) Duration() int { return p.duration }

func (p *Plan) Expenses() []Expense {
	res := make([]Expense, len(p.expenses))
	copy(res, p.expenses)
	return res
}

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

func (p *Plan) MonthlyExpense(month MonthIndex) Money {
	var total Money
	for _, exp := range p.expenses {
		if exp.Month == month {
			total += exp.Amount
		}
	}
	return total
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

func (p *Plan) MonthlyNetCashFlow(month MonthIndex) Money {
	return p.MonthlyRevenue(month) - p.MonthlyExpense(month)
}

func (p *Plan) TotalExpenses() Money {
	var total Money = 0
	for _, exp := range p.expenses {
		total += exp.Amount
	}
	return total
}

func (p *Plan) TotalRevenues() Money {
	var total Money = 0
	for _, exp := range p.revenues {
		total += exp.Amount
	}
	return total
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

// MonthlyDepreciation calculates the total non-cash depreciation expense for all assets in a given month.
func (p *Plan) MonthlyDepreciation(month MonthIndex) Money {
	var total Money
	for _, asset := range p.futurePurchases {
		total += asset.DepreciationForMonth(month)
	}
	return total
}
