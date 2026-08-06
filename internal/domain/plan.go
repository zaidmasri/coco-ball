// Package domain contains the aggregate root for the business plan
package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
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
	None            DepreciationMethod = "None"

	FlatGrowth        GrowthType = "Flat"
	AnnualStepPercent GrowthType = "AnnualStepPercent"
)

// Wizard hub keys. A "hub" is a top-level plan section (Starting Point,
// Payroll, Sales Forecast, Cash Flow) made up of one or more sub-sections,
// each tracked independently in the wizard_sections table (see
// internal/store's GetWizardSectionStatus/MarkWizardSectionComplete).
const (
	HubStartingPoint     = "starting-point"
	HubPayroll           = "payroll"
	HubSalesForecast     = "sales-forecast"
	HubOperatingExpenses = "operating-expenses"
	HubCashFlow          = "cash-flow"
)

// Wizard section keys. Shared by the store layer (section completion
// tracking), the handlers layer (wizard route path segments), and the
// views layer (summary page rendering) so all three always agree on the
// same literal strings.
const (
	SectionFixedAssets    = "fixed-assets"
	SectionStartupCosts   = "startup-costs"
	SectionFundingSources = "funding-sources"
	SectionCashOnHand     = "cash-on-hand"

	SectionSalaryRoles     = "salary-roles"
	SectionBenefits        = "benefits"
	SectionPayrollTaxRates = "payroll-tax-rates"

	SectionProducts         = "products"
	SectionSalesGrowthCurve = "sales-growth-curve"

	SectionInventoryPurchases = "inventory-purchases"
	SectionDistributions      = "distributions"

	// Operating Expenses is a single-section hub: the hub and its one
	// section share this key (stored in different wizard_sections
	// columns), since there's no multi-section summary page to route
	// through - it's a single repeatable list, on par with Fixed Assets/
	// Salary Roles/Products rather than a hub of several such lists.
	SectionOperatingExpenses = "operating-expenses"
)

var (
	ErrInvalidName                      = errors.New("name cannot be empty")
	ErrNegativeAmount                   = errors.New("expense amount cannot be negative")
	ErrInvalidMonthIndex                = errors.New("month index is out of bounds")
	ErrInvalidStartingMonth             = errors.New("plan starting month must be between 1-12")
	ErrInvalidDepreciationMethod        = errors.New("invalid deprecation method")
	ErrInvalidUsefulLife                = errors.New("invalid useful life")
	ErrPurchaseCostLessThanSalvageValue = errors.New("purchase cost cannot be less than salvage value")
	ErrInvalidGrowthType                = errors.New("invalid growth type")
	ErrInvalidStartingYear              = errors.New("plan starting year must be between 1900 and 2100")
)

// Plan is the aggregate root. Its fields are populated from two sources:
// the plans.data JSON blob (core details, revenue, opex, cogs - see
// plan_json.go) and a set of normalized SQL tables loaded by the store
// layer after unmarshalling (Starting Point, Payroll, Sales Forecast, Cash
// Flow - see each section's LoadXData method).
type Plan struct {
	id               uuid.UUID
	name             string
	startingMonth    int
	startingYear     int
	ownerID          uuid.UUID
	revenues         []RevenueStream
	opEx             []Cost
	cogs             []Cost
	futurePurchases  []CapitalAsset
	startupCosts     []StartupCost
	fundingSources   []FundingSource
	startingBalances StartingBalances
	salaryRoles      []SalaryRole
	benefits         []Benefit
	payrollTaxRates  PayrollTaxRates
	products         []Product
	salesGrowth      SalesGrowthCurve
	inventoryPlan    []InventoryPurchase
	distributions    []Distribution
}

func validateCoreProperties(name string, month, year int) (string, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return "", ErrInvalidName
	}

	if month < 1 || month > 12 {
		return "", ErrInvalidStartingMonth
	}

	// Add a sensible constraint for your business logic
	if year < 1900 || year > 2100 {
		return "", ErrInvalidStartingYear
	}

	return cleanName, nil
}

func NewPlan(id uuid.UUID, name string, startingMonth, startingYear int, ownerID uuid.UUID) (*Plan, error) {
	cleanName, err := validateCoreProperties(name, startingMonth, startingYear)
	if err != nil {
		return nil, err
	}

	return &Plan{
		id:            id,
		name:          cleanName,
		startingMonth: startingMonth,
		startingYear:  startingYear,
		ownerID:       ownerID,
		// Always initialize slices so they aren't nil
		revenues:        make([]RevenueStream, 0),
		opEx:            make([]Cost, 0),
		cogs:            make([]Cost, 0),
		futurePurchases: make([]CapitalAsset, 0),
		startupCosts:    make([]StartupCost, 0),
		fundingSources:  make([]FundingSource, 0),
		salaryRoles:     make([]SalaryRole, 0),
		benefits:        make([]Benefit, 0),
		products:        make([]Product, 0),
		inventoryPlan:   make([]InventoryPurchase, 0),
		distributions:   make([]Distribution, 0),
	}, nil
}

func (p *Plan) ChangeCoreDetails(name string, startingMonth, startingYear int) error {
	cleanName, err := validateCoreProperties(name, startingMonth, startingYear)
	if err != nil {
		return err // Reject the update entirely
	}

	// Only mutate state after all validations have passed
	p.name = cleanName
	p.startingMonth = startingMonth
	p.startingYear = startingYear

	return nil
}

func (p *Plan) ID() uuid.UUID      { return p.id }
func (p *Plan) Name() string       { return p.name }
func (p *Plan) StartingMonth() int { return p.startingMonth }
func (p *Plan) StartingYear() int  { return p.startingYear }
func (p *Plan) OwnerID() uuid.UUID { return p.ownerID }
