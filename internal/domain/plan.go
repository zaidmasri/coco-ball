// Package domain contains the aggregate root for the business plan
package domain

import (
	"errors"
	"strings"
)

type Money int64

// MonthIndex normalizes time. 0 = Month 1 of operations, 1 = Month 2, etc.
type MonthIndex int64

var (
	ErrInvalidName       = errors.New("name cannot be empty")
	ErrNegativeAmount    = errors.New("expense amount cannot be negative")
	ErrInvalidMonthIndex = errors.New("month index is out of bounds")
	ErrInvalidDuration   = errors.New("plan duration must be at least 1 month")
)

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
	id       int
	name     string
	duration int // Total months the projection spans
	expenses []Expense
	revenues []Revenue
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
		expenses: make([]Expense, 0),
		revenues: make([]Revenue, 0),
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
	result := make([]Expense, len(p.expenses))
	copy(result, p.expenses)
	return result
}

func (p *Plan) Revenues() []Revenue {
	result := make([]Revenue, len(p.revenues))
	copy(result, p.revenues)
	return result
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
