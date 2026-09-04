package entities

import "uuid"

// Product is a single sales-forecast product/service line. It carries its
// own identity (id) so a wizard row and the domain value it stores can be
// referenced by the same UUID - mirrors Cost (plan_growth.go) and
// SalaryRole (plan_payroll.go).
type Product struct {
	id           uuid.UUID
	Name         string
	Month1Units  int
	PricePerUnit Money
	CostPerUnit  Money
}

func (p Product) ID() uuid.UUID { return p.id }

// SetID overrides a Product's identity. Used only by repository
// implementations reconstructing a Product already persisted under a known
// ID (the wizard item's row ID) - mirrors SalaryRole.SetID. Reconstructed
// rows may be incomplete drafts, so this deliberately bypasses NewProduct's
// validation.
func (p *Product) SetID(id uuid.UUID) { p.id = id }

// NewProduct creates a new Product line item with a domain-generated UUIDv7
// identity, validating it against the same invariants a persisted Product
// must satisfy. Mirrors NewSalaryRole's shape.
func NewProduct(name string, month1Units int, pricePerUnit, costPerUnit Money) (Product, error) {
	p := Product{
		id:           uuid.NewV7(),
		Name:         name,
		Month1Units:  month1Units,
		PricePerUnit: pricePerUnit,
		CostPerUnit:  costPerUnit,
	}
	if err := ValidateProduct(p); err != nil {
		return Product{}, err
	}
	return p, nil
}

// ValidatedProduct is an opaque token proving a Product passed every
// invariant ValidateProduct checks. It can only be produced by
// NewValidatedProduct - mirrors ValidatedSalaryRole's shape.
// ProductRepository.CompleteProduct accepts only this type.
type ValidatedProduct struct {
	product     Product
	isValidated bool
}

// NewValidatedProduct validates an existing Product value - including one
// built while reconstructing a wizard draft - and wraps it.
func NewValidatedProduct(p Product) (ValidatedProduct, error) {
	if err := ValidateProduct(p); err != nil {
		return ValidatedProduct{}, err
	}
	return ValidatedProduct{product: p, isValidated: true}, nil
}

func (v ValidatedProduct) Product() Product { return v.product }

// SalesGrowthCurve is the global unit-sales growth schedule applied to every
// product line: quarterly rates for Year 1, then one rate per subsequent year.
type SalesGrowthCurve struct {
	Year1QuarterlyRates [4]float64
	FutureYearRates     []float64
}

// LoadSalesForecastData populates the Sales Forecast section fields from an
// external source. Mirrors LoadStartingPointData: Sales Forecast moved to
// normalized SQL tables, so Plan's own JSON (un)marshalling no longer
// carries this data.
func (p *Plan) LoadSalesForecastData(products []Product, growth SalesGrowthCurve) {
	p.products = products
	p.salesGrowth = growth
}

// ClearSalesForecast wipes existing sales forecast data.
func (p *Plan) ClearSalesForecast() {
	p.products = make([]Product, 0)
	p.salesGrowth = SalesGrowthCurve{}
}

// ValidateProduct checks a Product before it's persisted.
func ValidateProduct(product Product) error {
	if _, err := validateRequiredName(product.Name); err != nil {
		return err
	}
	if err := validateMoneyAmount(product.PricePerUnit); err != nil {
		return err
	}
	if err := validateMoneyAmount(product.CostPerUnit); err != nil {
		return err
	}
	return nil
}

// ValidateSalesGrowthCurve checks the global unit-sales growth schedule
// before the singleton Sales Growth Curve section is marked complete.
func ValidateSalesGrowthCurve(curve SalesGrowthCurve) error {
	for _, rate := range curve.Year1QuarterlyRates {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	for _, rate := range curve.FutureYearRates {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// AddProduct appends a product/service sales line.
func (p *Plan) AddProduct(product Product) error {
	if err := ValidateProduct(product); err != nil {
		return err
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
