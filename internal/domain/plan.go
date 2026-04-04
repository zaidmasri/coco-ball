// Package domain contains the aggregate root for the business plan
package domain

import (
	"errors"
	"strings"
)

type Money int64

var (
	ErrInvalidName    = errors.New("name cannot be empty")
	ErrNegativeAmount = errors.New("expense amount cannot be negative")
)

type Expense struct {
	Name   string
	Amount Money
}

type Plan struct {
	id       int
	name     string
	expenses []Expense
}

func NewPlan(id int, name string) (*Plan, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}

	return &Plan{
		id:   id,
		name: name,
		// Always initialize slices so they aren't nil
		expenses: make([]Expense, 0),
	}, nil
}

func (p *Plan) AddExpense(name string, amount Money) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if amount < 0 {
		return ErrNegativeAmount
	}

	p.expenses = append(p.expenses, Expense{
		Name:   name,
		Amount: amount,
	})
	return nil
}

func (p *Plan) ID() int      { return p.id }
func (p *Plan) Name() string { return p.name }

func (p *Plan) Expenses() []Expense {
	result := make([]Expense, len(p.expenses))
	copy(result, p.expenses)
	return result
}

func (p *Plan) TotalExpenses() Money {
	var total Money = 0
	for _, exp := range p.expenses {
		total += exp.Amount
	}
	return total
}
