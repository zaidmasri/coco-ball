// Package store handles persistence of system data
package store

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// PlanStore defines how we save and retrieve plans.
type PlanStore interface {
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
	Delete(id uuid.UUID) error

	// User management (basic)
	SaveUser(u *domain.User) error
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)

	// User management (with password)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)

	// Session management
	SaveSession(s *domain.Session) error
	GetSession(sessionID string) (*domain.Session, error)
	DeleteSession(sessionID string) error

	// Access control
	GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error
	GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error)
	GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error)
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)

	// Invites
	CreateInvite(invite *domain.PlanInvite) error
	GetInvite(id uuid.UUID) (*domain.PlanInvite, error)
	GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error)
	GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error)
	UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error

	// Starting Point: Fixed Assets wizard
	CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error)
	GetCapitalAssetDraft(planID uuid.UUID) (*CapitalAssetItem, error)
	GetCapitalAsset(itemID uuid.UUID) (*CapitalAssetItem, error)
	SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status ItemStatus) error
	ListCompleteCapitalAssets(planID uuid.UUID) ([]CapitalAssetItem, error)
	DeleteCapitalAsset(itemID uuid.UUID) error

	// Starting Point: Startup Costs wizard
	CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error)
	GetStartupCostDraft(planID uuid.UUID) (*StartupCostItem, error)
	GetStartupCost(itemID uuid.UUID) (*StartupCostItem, error)
	SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status ItemStatus) error
	ListCompleteStartupCosts(planID uuid.UUID) ([]StartupCostItem, error)
	DeleteStartupCost(itemID uuid.UUID) error

	// Starting Point: Funding Sources wizard
	CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error)
	GetFundingSourceDraft(planID uuid.UUID) (*FundingSourceItem, error)
	GetFundingSource(itemID uuid.UUID) (*FundingSourceItem, error)
	SaveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status ItemStatus) error
	ListCompleteFundingSources(planID uuid.UUID) ([]FundingSourceItem, error)
	DeleteFundingSource(itemID uuid.UUID) error

	// Starting Point: Cash on Hand wizard (singleton per plan)
	GetStartingBalancesRow(planID uuid.UUID) (*StartingBalancesRow, error)
	SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error

	// Payroll: Salary Roles wizard
	CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error)
	GetSalaryRoleDraft(planID uuid.UUID) (*SalaryRoleItem, error)
	GetSalaryRole(itemID uuid.UUID) (*SalaryRoleItem, error)
	SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status ItemStatus) error
	ListCompleteSalaryRoles(planID uuid.UUID) ([]SalaryRoleItem, error)
	DeleteSalaryRole(itemID uuid.UUID) error

	// Payroll: Benefits wizard
	CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error)
	GetBenefitDraft(planID uuid.UUID) (*BenefitItem, error)
	GetBenefit(itemID uuid.UUID) (*BenefitItem, error)
	SaveBenefitStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int, status ItemStatus) error
	ListCompleteBenefits(planID uuid.UUID) ([]BenefitItem, error)
	DeleteBenefit(itemID uuid.UUID) error

	// Payroll: Tax Rates wizard (singleton per plan)
	GetPayrollTaxRatesRow(planID uuid.UUID) (*PayrollTaxRatesRow, error)
	SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error

	// Sales Forecast: Products wizard
	CreateProductDraft(planID uuid.UUID) (uuid.UUID, error)
	GetProductDraft(planID uuid.UUID) (*ProductItem, error)
	GetProduct(itemID uuid.UUID) (*ProductItem, error)
	SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status ItemStatus) error
	ListCompleteProducts(planID uuid.UUID) ([]ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error

	// Sales Forecast: Sales Growth Curve wizard (singleton per plan)
	GetSalesGrowthCurveRow(planID uuid.UUID) (*SalesGrowthCurveRow, error)
	SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error

	// Cash Flow: Inventory Purchases wizard
	CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error)
	GetInventoryPurchaseDraft(planID uuid.UUID) (*InventoryPurchaseItem, error)
	GetInventoryPurchase(itemID uuid.UUID) (*InventoryPurchaseItem, error)
	SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status ItemStatus) error
	ListCompleteInventoryPurchases(planID uuid.UUID) ([]InventoryPurchaseItem, error)
	DeleteInventoryPurchase(itemID uuid.UUID) error

	// Cash Flow: Distributions wizard
	CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error)
	GetDistributionDraft(planID uuid.UUID) (*DistributionItem, error)
	GetDistribution(itemID uuid.UUID) (*DistributionItem, error)
	SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status ItemStatus) error
	ListCompleteDistributions(planID uuid.UUID) ([]DistributionItem, error)
	DeleteDistribution(itemID uuid.UUID) error

	// Operating Expenses wizard
	CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error)
	GetOperatingExpenseDraft(planID uuid.UUID) (*OperatingExpenseItem, error)
	GetOperatingExpense(itemID uuid.UUID) (*OperatingExpenseItem, error)
	SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status ItemStatus) error
	ListCompleteOperatingExpenses(planID uuid.UUID) ([]OperatingExpenseItem, error)
	DeleteOperatingExpense(itemID uuid.UUID) error

	// Wizard section-level explicit completion, generalized across all
	// five hubs (Starting Point, Payroll, Sales Forecast, Operating
	// Expenses, Cash Flow) via a hub key, independent of item counts
	// (deleting a section's last item does not un-complete it).
	MarkWizardSectionComplete(planID uuid.UUID, hub, section string) error
	GetWizardSectionStatus(planID uuid.UUID, hub string) (map[string]bool, error)
}

// ItemStatus tracks whether a wizard item has been fully answered
// ("complete", and therefore safe to include in financial projections) or
// is still mid-flow ("draft").
type ItemStatus string

const (
	StatusDraft    ItemStatus = "draft"
	StatusComplete ItemStatus = "complete"
)

// CapitalAssetItem is a Fixed Assets wizard row: a domain.CapitalAsset plus
// the wizard-progress metadata the domain type itself has no business
// knowing about.
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

// SalesGrowthCurveRow is the Sales Forecast "Sales Growth Curve" singleton
// row for a plan.
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
