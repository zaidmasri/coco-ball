package entities

import "uuid"

// InventoryPurchase is a discretionary additional-inventory cash outflow. It
// carries its own identity (id) so a wizard row and the domain value it
// stores can be referenced by the same UUID - mirrors Cost (plan_growth.go)
// and SalaryRole/Benefit (plan_payroll.go).
type InventoryPurchase struct {
	id             uuid.UUID
	Category       string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

func (i InventoryPurchase) ID() uuid.UUID { return i.id }

// SetID overrides an InventoryPurchase's identity. Used only by repository
// implementations reconstructing an InventoryPurchase already persisted
// under a known ID (the wizard item's row ID) - mirrors SalaryRole.SetID.
// Reconstructed rows may be incomplete drafts, so this deliberately bypasses
// NewInventoryPurchase's validation.
func (i *InventoryPurchase) SetID(id uuid.UUID) { i.id = id }

// NewInventoryPurchase creates a new InventoryPurchase line item with a
// domain-generated UUIDv7 identity, validating it against the same
// invariants a persisted InventoryPurchase must satisfy. Mirrors
// NewSalaryRole's shape.
func NewInventoryPurchase(category string, monthlyAmount Money, growth AnnualGrowth) (InventoryPurchase, error) {
	i := InventoryPurchase{
		id:             uuid.NewV7(),
		Category:       category,
		MonthlyAmount:  monthlyAmount,
		GrowthAfterYr1: growth,
	}
	if err := ValidateInventoryPurchase(i); err != nil {
		return InventoryPurchase{}, err
	}
	return i, nil
}

// ValidatedInventoryPurchase is an opaque token proving an InventoryPurchase
// passed every invariant ValidateInventoryPurchase checks. It can only be
// produced by NewValidatedInventoryPurchase - mirrors ValidatedSalaryRole's
// shape (plan_payroll.go). InventoryPurchaseRepository.CompleteInventoryPurchase
// accepts only this type.
type ValidatedInventoryPurchase struct {
	purchase    InventoryPurchase
	isValidated bool
}

// NewValidatedInventoryPurchase validates an existing InventoryPurchase
// value - including one built while reconstructing a wizard draft - and
// wraps it.
func NewValidatedInventoryPurchase(i InventoryPurchase) (ValidatedInventoryPurchase, error) {
	if err := ValidateInventoryPurchase(i); err != nil {
		return ValidatedInventoryPurchase{}, err
	}
	return ValidatedInventoryPurchase{purchase: i, isValidated: true}, nil
}

func (v ValidatedInventoryPurchase) InventoryPurchase() InventoryPurchase { return v.purchase }

// Distribution is a discretionary owner distribution / repayment cash
// outflow. It carries its own identity (id) - mirrors InventoryPurchase/
// SalaryRole/Cost.
type Distribution struct {
	id             uuid.UUID
	Name           string
	MonthlyAmount  Money
	GrowthAfterYr1 AnnualGrowth
}

func (d Distribution) ID() uuid.UUID { return d.id }

// SetID overrides a Distribution's identity - mirrors InventoryPurchase.SetID.
func (d *Distribution) SetID(id uuid.UUID) { d.id = id }

// NewDistribution creates a new Distribution line item with a
// domain-generated UUIDv7 identity, validating it against the same
// invariants a persisted Distribution must satisfy. Mirrors NewSalaryRole's
// shape.
func NewDistribution(name string, monthlyAmount Money, growth AnnualGrowth) (Distribution, error) {
	d := Distribution{
		id:             uuid.NewV7(),
		Name:           name,
		MonthlyAmount:  monthlyAmount,
		GrowthAfterYr1: growth,
	}
	if err := ValidateDistribution(d); err != nil {
		return Distribution{}, err
	}
	return d, nil
}

// ValidatedDistribution is an opaque token proving a Distribution passed
// every invariant ValidateDistribution checks. It can only be produced by
// NewValidatedDistribution - mirrors ValidatedSalaryRole's shape.
// DistributionRepository.CompleteDistribution accepts only this type.
type ValidatedDistribution struct {
	distribution Distribution
	isValidated  bool
}

// NewValidatedDistribution validates an existing Distribution value and
// wraps it.
func NewValidatedDistribution(d Distribution) (ValidatedDistribution, error) {
	if err := ValidateDistribution(d); err != nil {
		return ValidatedDistribution{}, err
	}
	return ValidatedDistribution{distribution: d, isValidated: true}, nil
}

func (v ValidatedDistribution) Distribution() Distribution { return v.distribution }

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

// LoadCashFlowData populates the discretionary Cash Flow section fields
// from an external source. Mirrors LoadStartingPointData: Cash Flow moved
// to normalized SQL tables, so Plan's own JSON (un)marshalling no longer
// carries this data.
func (p *Plan) LoadCashFlowData(inventory []InventoryPurchase, distributions []Distribution) {
	p.inventoryPlan = inventory
	p.distributions = distributions
}

// ClearCashFlow wipes existing discretionary cash flow data.
func (p *Plan) ClearCashFlow() {
	p.inventoryPlan = make([]InventoryPurchase, 0)
	p.distributions = make([]Distribution, 0)
}

// ValidateInventoryPurchase checks an InventoryPurchase before it's persisted.
func ValidateInventoryPurchase(inv InventoryPurchase) error {
	if _, err := validateRequiredName(inv.Category); err != nil {
		return err
	}
	if err := validateMoneyAmount(inv.MonthlyAmount); err != nil {
		return err
	}
	for _, rate := range inv.GrowthAfterYr1.RatesAfterYear1 {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// AddInventoryPurchase appends an additional-inventory cash outflow line.
func (p *Plan) AddInventoryPurchase(inv InventoryPurchase) error {
	if err := ValidateInventoryPurchase(inv); err != nil {
		return err
	}
	p.inventoryPlan = append(p.inventoryPlan, inv)
	return nil
}

// ValidateDistribution checks a Distribution before it's persisted.
func ValidateDistribution(dist Distribution) error {
	if _, err := validateRequiredName(dist.Name); err != nil {
		return err
	}
	if err := validateMoneyAmount(dist.MonthlyAmount); err != nil {
		return err
	}
	for _, rate := range dist.GrowthAfterYr1.RatesAfterYear1 {
		if err := validateGrowthRate(rate); err != nil {
			return err
		}
	}
	return nil
}

// AddDistribution appends an owner distribution / repayment cash outflow line.
func (p *Plan) AddDistribution(dist Distribution) error {
	if err := ValidateDistribution(dist); err != nil {
		return err
	}
	p.distributions = append(p.distributions, dist)
	return nil
}

func (p *Plan) AdditionalInventory() []InventoryPurchase { return p.inventoryPlan }
func (p *Plan) Distributions() []Distribution            { return p.distributions }
