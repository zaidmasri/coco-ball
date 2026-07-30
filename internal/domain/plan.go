// Package domain contains the aggregate root for the business plan
package domain

import (
	"encoding/json"
	"errors"
	"math"
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

// SalaryRole is a single payroll line item (a role/title with headcount).
type SalaryRole struct {
	Role           string
	IsContractor   bool
	Headcount      int
	MonthlyPay     Money
	GrowthAfterYr1 AnnualGrowth
}

// Benefit is a single employee benefit line item.
type Benefit struct {
	Type           string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

// PayrollTaxRates holds employer-side payroll tax rates.
type PayrollTaxRates struct {
	SocialSecurityRate float64
	MedicareRate       float64
	FUTARate           float64
	SUTARate           float64
}

// Product is a single sales-forecast product/service line.
type Product struct {
	Name         string
	Month1Units  int
	PricePerUnit Money
	CostPerUnit  Money
}

// SalesGrowthCurve is the global unit-sales growth schedule applied to every
// product line: quarterly rates for Year 1, then one rate per subsequent year.
type SalesGrowthCurve struct {
	Year1QuarterlyRates [4]float64
	FutureYearRates     []float64
}

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
			// Compound the base amount annually: Base * (1 + rate)^years
			multiplier := math.Pow(1.0+c.Growth.AnnualRate, float64(yearsPassed))
			return Money(float64(c.BaseAmountPerMonth) * multiplier)
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
			return Money(float64(r.BaseAmountPerMonth) * multiplier)
		}
	}
	return r.BaseAmountPerMonth
}

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

// ClearStartingPoint wipes the existing starting point data.
// This is critical for dynamic forms so we don't duplicate data when editing.
func (p *Plan) ClearStartingPoint() {
	p.futurePurchases = make([]CapitalAsset, 0)
	p.startupCosts = make([]StartupCost, 0)
	p.fundingSources = make([]FundingSource, 0)
	p.startingBalances = StartingBalances{}
}

func (p *Plan) AddStartupCost(name string, amount Money) {
	if strings.TrimSpace(name) != "" && amount > 0 {
		p.startupCosts = append(p.startupCosts, StartupCost{Name: name, Amount: amount})
	}
}

func (p *Plan) AddFundingSource(name string, amount Money, rate float64, term int) {
	if strings.TrimSpace(name) != "" && amount > 0 {
		p.fundingSources = append(p.fundingSources, FundingSource{
			Name: name, Amount: amount, InterestRate: rate, TermMonths: term,
		})
	}
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

// ClearPayroll wipes existing payroll data so forms can cleanly overwrite it.
func (p *Plan) ClearPayroll() {
	p.salaryRoles = make([]SalaryRole, 0)
	p.benefits = make([]Benefit, 0)
	p.payrollTaxRates = PayrollTaxRates{}
}

// AddSalaryRole appends a salary/wage line item.
func (p *Plan) AddSalaryRole(role SalaryRole) error {
	if strings.TrimSpace(role.Role) == "" {
		return ErrInvalidName
	}
	if role.MonthlyPay < 0 {
		return ErrNegativeAmount
	}
	p.salaryRoles = append(p.salaryRoles, role)
	return nil
}

// AddBenefit appends an employee benefit line item.
func (p *Plan) AddBenefit(benefit Benefit) error {
	if strings.TrimSpace(benefit.Type) == "" {
		return ErrInvalidName
	}
	if benefit.MonthlyAmount < 0 {
		return ErrNegativeAmount
	}
	p.benefits = append(p.benefits, benefit)
	return nil
}

// SetPayrollTaxRates sets the employer payroll tax rates.
func (p *Plan) SetPayrollTaxRates(rates PayrollTaxRates) {
	p.payrollTaxRates = rates
}

func (p *Plan) SalaryRoles() []SalaryRole { return p.salaryRoles }
func (p *Plan) Benefits() []Benefit       { return p.benefits }
func (p *Plan) PayrollTaxRates() PayrollTaxRates {
	return p.payrollTaxRates
}

// HasPayrollTaxRates reports whether any payroll tax rate has been set,
// used by templates to decide whether to show suggested defaults.
func (r PayrollTaxRates) HasPayrollTaxRates() bool {
	return r.SocialSecurityRate != 0 || r.MedicareRate != 0 || r.FUTARate != 0 || r.SUTARate != 0
}

func (r PayrollTaxRates) SocialSecurityRatePercent() float64 { return r.SocialSecurityRate * 100 }
func (r PayrollTaxRates) MedicareRatePercent() float64       { return r.MedicareRate * 100 }
func (r PayrollTaxRates) FUTARatePercent() float64           { return r.FUTARate * 100 }
func (r PayrollTaxRates) SUTARatePercent() float64           { return r.SUTARate * 100 }

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (s SalaryRole) GrowthRatePercent(index int) float64 {
	return s.GrowthAfterYr1.GrowthRatePercent(index)
}

// GrowthRatePercent returns the growth rate at the given index as a
// percentage, for use directly in HTML templates.
func (b Benefit) GrowthRatePercent(index int) float64 {
	return b.GrowthAfterYr1.GrowthRatePercent(index)
}

// ClearSalesForecast wipes existing sales forecast data.
func (p *Plan) ClearSalesForecast() {
	p.products = make([]Product, 0)
	p.salesGrowth = SalesGrowthCurve{}
}

// AddProduct appends a product/service sales line.
func (p *Plan) AddProduct(product Product) error {
	if strings.TrimSpace(product.Name) == "" {
		return ErrInvalidName
	}
	if product.PricePerUnit < 0 || product.CostPerUnit < 0 {
		return ErrNegativeAmount
	}
	p.products = append(p.products, product)
	return nil
}

// SetSalesGrowth sets the global unit-sales growth curve.
func (p *Plan) SetSalesGrowth(curve SalesGrowthCurve) {
	p.salesGrowth = curve
}

func (p *Plan) Products() []Product           { return p.products }
func (p *Plan) SalesGrowth() SalesGrowthCurve { return p.salesGrowth }

// Year1QuarterlyPercent returns the Year 1 quarterly growth rate at the given
// index (0-3) as a percentage, for use directly in HTML templates.
func (s SalesGrowthCurve) Year1QuarterlyPercent(index int) float64 {
	if index < 0 || index >= len(s.Year1QuarterlyRates) {
		return 0
	}
	return s.Year1QuarterlyRates[index] * 100
}

// FutureYearPercent returns the future-year growth rate at the given index as
// a percentage, for use directly in HTML templates.
func (s SalesGrowthCurve) FutureYearPercent(index int) float64 {
	if index < 0 || index >= len(s.FutureYearRates) {
		return 0
	}
	return s.FutureYearRates[index] * 100
}

// ClearCashFlow wipes existing discretionary cash flow data.
func (p *Plan) ClearCashFlow() {
	p.inventoryPlan = make([]InventoryPurchase, 0)
	p.distributions = make([]Distribution, 0)
}

// AddInventoryPurchase appends an additional-inventory cash outflow line.
func (p *Plan) AddInventoryPurchase(inv InventoryPurchase) error {
	if strings.TrimSpace(inv.Category) == "" {
		return ErrInvalidName
	}
	if inv.MonthlyAmount < 0 {
		return ErrNegativeAmount
	}
	p.inventoryPlan = append(p.inventoryPlan, inv)
	return nil
}

// AddDistribution appends an owner distribution / repayment cash outflow line.
func (p *Plan) AddDistribution(dist Distribution) error {
	if strings.TrimSpace(dist.Name) == "" {
		return ErrInvalidName
	}
	if dist.MonthlyAmount < 0 {
		return ErrNegativeAmount
	}
	p.distributions = append(p.distributions, dist)
	return nil
}

func (p *Plan) AdditionalInventory() []InventoryPurchase { return p.inventoryPlan }
func (p *Plan) Distributions() []Distribution            { return p.distributions }

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

func (p *Plan) ID() uuid.UUID                      { return p.id }
func (p *Plan) Name() string                       { return p.name }
func (p *Plan) StartingMonth() int                 { return p.startingMonth }
func (p *Plan) StartingYear() int                  { return p.startingYear }
func (p *Plan) OwnerID() uuid.UUID                 { return p.ownerID }
func (p *Plan) StartupCosts() []StartupCost        { return p.startupCosts }
func (p *Plan) FundingSources() []FundingSource    { return p.fundingSources }
func (p *Plan) StartingBalances() StartingBalances { return p.startingBalances }

// UsefulLifeYears formatting Helpers for the HTML Templates ---
// The form takes Years, but the domain stores Months. This converts it back for the form.
func (c CapitalAsset) UsefulLifeYears() int {
	return c.UsefulLifeMonths / 12
}

// The form takes 5.5%, but the domain stores 0.055. This converts it back for the form.
func (f FundingSource) InterestRatePercent() float64 {
	return f.InterestRate * 100
}

func (p *Plan) Revenues() []RevenueStream {
	res := make([]RevenueStream, len(p.revenues))
	copy(res, p.revenues)
	return res
}

func (p *Plan) FuturePurchases() []CapitalAsset {
	res := make([]CapitalAsset, len(p.futurePurchases))
	copy(res, p.futurePurchases)
	return res
}

// MonthlyRevenue calculates the projected revenue for a specific month
func (p *Plan) MonthlyRevenue(month MonthIndex) Money {
	var total Money
	for _, rev := range p.revenues {
		total += rev.ProjectedAmount(month)
	}
	return total
}

// TotalRevenues now loops through the timeline to sum the continuous curves.
func (p *Plan) TotalRevenues(duration int) Money {
	var total Money
	for i := range duration {
		total += p.MonthlyRevenue(MonthIndex(i))
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

func (p *Plan) MonthlyNetCashFlow(month MonthIndex, duration int) Money {
	if int(month) < 0 || int(month) >= duration {
		return 0
	}
	return p.MonthlyRevenue(month) - p.MonthlyOpEx(month) - p.MonthlyCOGS(month)
}

// TotalOpEx calculates the lifetime operating expenses over the plan's duration.
func (p *Plan) TotalOpEx(duration int) Money {
	var total Money
	for i := range duration {
		total += p.MonthlyOpEx(MonthIndex(i))
	}
	return total
}

// TotalCOGS calculates the lifetime COGS over the plan's duration.
func (p *Plan) TotalCOGS(duration int) Money {
	var total Money
	for i := range duration {
		total += p.MonthlyCOGS(MonthIndex(i))
	}
	return total
}

// TotalExpenses calculates the total lifetime expenses (OpEx + COGS).
func (p *Plan) TotalExpenses(duration int) Money {
	return p.TotalOpEx(duration) + p.TotalCOGS(duration)
}

func (p *Plan) AddRevenue(name string, baseAmount Money, growth GrowthStrategy) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	if baseAmount < 0 {
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
	if asset.DepreciationMethod != None && asset.UsefulLifeMonths < 1 {
		return ErrInvalidUsefulLife
	}
	if asset.PurchaseCost < asset.SalvageValue {
		return ErrPurchaseCostLessThanSalvageValue
	}

	if asset.DepreciationMethod != StraightLine && asset.DepreciationMethod != DoubleDeclining && asset.DepreciationMethod != None {
		return ErrInvalidDepreciationMethod
	}

	p.futurePurchases = append(p.futurePurchases, asset)
	return nil
}

// planJSON is an intermediate struct for JSON marshalling
type planJSON struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	StartingMonth       int                 `json:"startingMonth"`
	StartingYear        int                 `json:"startingYear"`
	OwnerID             string              `json:"ownerID"`
	Revenues            []RevenueStream     `json:"revenues"`
	OpEx                []Cost              `json:"opEx"`
	COGS                []Cost              `json:"cogs"`
	FuturePurchases     []CapitalAsset      `json:"futurePurchases"`
	StartupCosts        []StartupCost       `json:"startupCosts"`
	FundingSources      []FundingSource     `json:"fundingSources"`
	StartingBalances    StartingBalances    `json:"startingBalances"`
	SalaryRoles         []SalaryRole        `json:"salaryRoles"`
	Benefits            []Benefit           `json:"benefits"`
	PayrollTaxRates     PayrollTaxRates     `json:"payrollTaxRates"`
	Products            []Product           `json:"products"`
	SalesGrowth         SalesGrowthCurve    `json:"salesGrowth"`
	AdditionalInventory []InventoryPurchase `json:"additionalInventory"`
	Distributions       []Distribution      `json:"distributions"`
}

// MarshalJSON implements json.Marshaler for Plan
func (p *Plan) MarshalJSON() ([]byte, error) {
	pj := planJSON{
		ID:                  p.id.String(),
		Name:                p.name,
		StartingMonth:       p.startingMonth,
		StartingYear:        p.startingYear,
		OwnerID:             p.ownerID.String(),
		Revenues:            p.revenues,
		OpEx:                p.opEx,
		COGS:                p.cogs,
		FuturePurchases:     p.futurePurchases,
		StartupCosts:        p.startupCosts,
		FundingSources:      p.fundingSources,
		StartingBalances:    p.startingBalances,
		SalaryRoles:         p.salaryRoles,
		Benefits:            p.benefits,
		PayrollTaxRates:     p.payrollTaxRates,
		Products:            p.products,
		SalesGrowth:         p.salesGrowth,
		AdditionalInventory: p.inventoryPlan,
		Distributions:       p.distributions,
	}

	return json.Marshal(pj)
}

// UnmarshalJSON implements json.Unmarshaler for Plan
func (p *Plan) UnmarshalJSON(data []byte) error {
	var pj planJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return err
	}

	id, err := uuid.Parse(pj.ID)
	if err != nil {
		return err
	}

	ownerID, err := uuid.Parse(pj.OwnerID)
	if err != nil {
		return err
	}

	p.id = id
	p.name = pj.Name
	p.startingMonth = pj.StartingMonth
	p.startingYear = pj.StartingYear
	p.ownerID = ownerID
	p.revenues = pj.Revenues
	p.opEx = pj.OpEx
	p.cogs = pj.COGS
	p.futurePurchases = pj.FuturePurchases
	p.startupCosts = pj.StartupCosts
	p.fundingSources = pj.FundingSources
	p.startingBalances = pj.StartingBalances
	p.salaryRoles = pj.SalaryRoles
	p.benefits = pj.Benefits
	p.payrollTaxRates = pj.PayrollTaxRates
	p.products = pj.Products
	p.salesGrowth = pj.SalesGrowth
	p.inventoryPlan = pj.AdditionalInventory
	p.distributions = pj.Distributions

	return nil
}
