// wizard_item.go defines the item-wrapper types used by the wizard sub-entity
// repositories. Each *Item struct pairs a domain value object (e.g.
// domain.CapitalAsset) with a store-assigned UUID and wizard progress fields.
//
// NOTE: These IDs live at the repository layer, not the domain layer. A
// future improvement is to promote the ID into the domain entity itself so
// that Plan can reference individual items by identity without the repository
// wrapper — see AGENTS.md "Known Limitations" for context.
package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// ItemStatus tracks whether a wizard item has been fully answered
// ("complete", and therefore safe to include in financial projections) or
// is still mid-flow ("draft").
type ItemStatus string

const (
	StatusDraft    ItemStatus = "draft"
	StatusComplete ItemStatus = "complete"
)

// CapitalAssetItem is a Fixed Assets wizard row.
type CapitalAssetItem struct {
	ID          uuid.UUID
	Asset       domain.CapitalAsset
	Status      ItemStatus
	CurrentStep int
}

// StartupCostItem is a Startup Costs wizard row.
type StartupCostItem struct {
	ID          uuid.UUID
	Cost        domain.StartupCost
	Status      ItemStatus
	CurrentStep int
}

// FundingSourceItem is a Funding Sources wizard row.
type FundingSourceItem struct {
	ID          uuid.UUID
	Funding     domain.FundingSource
	Status      ItemStatus
	CurrentStep int
}

// StartingBalancesRow is the Cash on Hand singleton row for a plan.
type StartingBalancesRow struct {
	Balances    domain.StartingBalances
	CurrentStep int
}

// SalaryRoleItem is a Payroll "Salary Roles" wizard row.
type SalaryRoleItem struct {
	ID          uuid.UUID
	Role        domain.SalaryRole
	Status      ItemStatus
	CurrentStep int
}

// BenefitItem is a Payroll "Benefits" wizard row.
type BenefitItem struct {
	ID          uuid.UUID
	Benefit     domain.Benefit
	Status      ItemStatus
	CurrentStep int
}

// PayrollTaxRatesRow is the Payroll "Tax Rates" singleton row for a plan.
type PayrollTaxRatesRow struct {
	Rates       domain.PayrollTaxRates
	CurrentStep int
}

// ProductItem is a Sales Forecast "Products" wizard row.
type ProductItem struct {
	ID          uuid.UUID
	Product     domain.Product
	Status      ItemStatus
	CurrentStep int
}

// SalesGrowthCurveRow is the Sales Forecast "Sales Growth Curve" singleton row for a plan.
type SalesGrowthCurveRow struct {
	Curve       domain.SalesGrowthCurve
	CurrentStep int
}

// InventoryPurchaseItem is a Cash Flow "Inventory Purchases" wizard row.
type InventoryPurchaseItem struct {
	ID          uuid.UUID
	Purchase    domain.InventoryPurchase
	Status      ItemStatus
	CurrentStep int
}

// DistributionItem is a Cash Flow "Distributions" wizard row.
type DistributionItem struct {
	ID           uuid.UUID
	Distribution domain.Distribution
	Status       ItemStatus
	CurrentStep  int
}

// OperatingExpenseItem is an Operating Expenses wizard row.
type OperatingExpenseItem struct {
	ID          uuid.UUID
	Cost        domain.Cost
	Status      ItemStatus
	CurrentStep int
}
