// Package repositories defines the persistence interfaces for each aggregate
// root and its child entities.
//
// # Aggregate root repositories
//
// PlanRepository and UserRepository are the only TRUE aggregate root
// repositories. They are the canonical persistence contract for their
// respective aggregates and are the interfaces that application services
// depend on.
//
// # Sub-entity repositories (wizard items)
//
// CapitalAssetRepository, SalaryRoleRepository, BenefitRepository,
// ProductRepository, InventoryPurchaseRepository, DistributionRepository,
// OperatingExpenseRepository, StartupCostRepository, FundingSourceRepository
// and the singleton-row helpers (StartingBalancesRepository,
// PayrollTaxRatesRepository, SalesGrowthCurveRepository) are NOT aggregate
// root repositories. They are convenience interfaces for the wizard's
// step-by-step persistence within the Plan aggregate boundary. All of the
// entities they manage belong to Plan; they exist separately only to allow
// progressive, per-section saves during the wizard flow.
package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// PlanRepository is the aggregate-root repository for Plan. It is the only
// interface that application services should depend on for plan persistence.
// Wizard-item repositories (CapitalAssetRepository, etc.) are sub-entity
// helpers and are injected into lower-level services, not the core plan service.
type PlanRepository interface {
	// Save persists a validated plan. The repository must also drain the
	// plan's domain events (via Plan.PullEvents) and write them to the
	// outbox table in the same transaction.
	Save(p domain.ValidatedPlan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
	Delete(id uuid.UUID) error
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)
}
