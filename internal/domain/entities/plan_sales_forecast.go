package entities

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
