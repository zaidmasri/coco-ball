package entities

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
)

type (
	// MonthIndex normalizes time. 0 = Month 1 of operations, 1 = Month 2, etc.
	MonthIndex         int64
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
// internal/infrastructure/sqlite's WizardProgressRepository).
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

	// domainEvents holds events accumulated during this request. The
	// repository must drain this list via PullEvents() and persist each event
	// to the outbox table in the same transaction that saves the plan.
	domainEvents []domevents.DomainEvent
}

// recordEvent appends a domain event to the plan's in-memory event list.
func (p *Plan) recordEvent(e domevents.DomainEvent) {
	p.domainEvents = append(p.domainEvents, e)
}

// PullEvents returns all accumulated domain events and resets the list.
// Call this inside the repository's Save implementation after persisting the
// plan, passing every event to the outbox writer in the same transaction.
func (p *Plan) PullEvents() []domevents.DomainEvent {
	evts := p.domainEvents
	p.domainEvents = nil
	return evts
}

// RecordUserInvited records a UserInvitedToPlan event on the Plan aggregate.
// Call this after creating the PlanInvite so the event is persisted in the
// same outbox transaction as the Plan save. PlanInvite is not an aggregate
// root and must not emit events directly.
func (p *Plan) RecordUserInvited(inviteID uuid.UUID, email string, level AccessLevel, invitedBy uuid.UUID) {
	p.recordEvent(domevents.NewUserInvitedToPlan(p.id, inviteID, invitedBy, email, string(level)))
}

// validate checks the plan's current state against all business invariants.
// It is separate from construction so it can be called independently.
func (p *Plan) validate() error {
	if strings.TrimSpace(p.name) == "" {
		return ErrInvalidName
	}
	if p.startingMonth < 1 || p.startingMonth > 12 {
		return ErrInvalidStartingMonth
	}
	if p.startingYear < 1900 || p.startingYear > 2100 {
		return ErrInvalidStartingYear
	}
	return nil
}

// ValidatedPlan is an opaque token proving the plan passed all invariant
// checks. Repository Save methods require this type so unvalidated data
// cannot be persisted — invalid states become unrepresentable at the
// persistence boundary.
type ValidatedPlan struct{ plan *Plan }

// Validate runs all invariant checks. On success it returns a ValidatedPlan
// that the repository layer accepts. On failure it returns a domain sentinel
// error describing the violation.
func (p *Plan) Validate() (ValidatedPlan, error) {
	if err := p.validate(); err != nil {
		return ValidatedPlan{}, err
	}
	return ValidatedPlan{plan: p}, nil
}

// Plan returns the underlying plan. Used only by the repository
// implementation to access plan data for persistence.
func (vp ValidatedPlan) Plan() *Plan { return vp.plan }

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

// NewPlan creates a brand-new Plan with a domain-generated UUIDv7 identity.
// owner must be a VerifiedUser (see user.go) — proof that the owner is a
// real, existing User of the system, not just any uuid.UUID a caller
// supplies. Use UnmarshalJSON (via plan_json.go) when reconstructing a plan
// from persisted data — that path sets the existing id/ownerID directly
// without calling this constructor, satisfying the rule that validation is
// write-side only.
func NewPlan(name string, startingMonth, startingYear int, owner VerifiedUser) (*Plan, error) {
	cleanName, err := validateCoreProperties(name, startingMonth, startingYear)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan id: %w", err)
	}

	p := &Plan{
		id:            id,
		name:          cleanName,
		startingMonth: startingMonth,
		startingYear:  startingYear,
		ownerID:       owner.ID(),
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
	}

	p.recordEvent(domevents.NewPlanCreated(id, cleanName, owner.ID()))
	return p, nil
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

	p.recordEvent(domevents.NewPlanUpdated(p.id))
	return nil
}

func (p *Plan) ID() uuid.UUID      { return p.id }
func (p *Plan) Name() string       { return p.name }
func (p *Plan) StartingMonth() int { return p.startingMonth }
func (p *Plan) StartingYear() int  { return p.startingYear }
func (p *Plan) OwnerID() uuid.UUID { return p.ownerID }

// AggregateID implements entities.AggregateRoot.
//
// Plan is the aggregate root for the business planning domain. Its boundary
// includes all wizard sections: capital assets, startup costs, funding
// sources, salary roles, benefits, products, inventory purchases,
// distributions, and operating expenses. These are entities within this
// aggregate and MUST NOT be referenced by other aggregates by embedding the
// full struct — reference this aggregate only by its ID (uuid.UUID).
//
// The only other aggregate root is User (see user.go). All cross-aggregate
// associations in this domain use uuid.UUID identifiers, never embedded structs.
func (p *Plan) AggregateID() uuid.UUID { return p.id }

// compile-time assertion that Plan satisfies the AggregateRoot marker.
var _ AggregateRoot = (*Plan)(nil)
