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

	// Starting Point: section-level explicit completion, independent of
	// item counts (deleting a section's last item does not un-complete it)
	MarkStartingPointSectionComplete(planID uuid.UUID, section string) error
	GetStartingPointSectionStatus(planID uuid.UUID) (map[string]bool, error)
}

// ItemStatus tracks whether a Starting Point wizard item has been fully
// answered ("complete", and therefore safe to include in financial
// projections) or is still mid-flow ("draft").
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
