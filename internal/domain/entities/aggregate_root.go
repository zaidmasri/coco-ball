// Package entities holds the domain entity types and value objects for the
// NorthBasis business planning domain.
//
// # Aggregate roots
//
// An aggregate root is the single entry point to a cluster of related domain
// objects. External code must reference a different aggregate only by its ID
// (uuid.UUID) — never by embedding the full aggregate struct. This prevents
// accidental coupling between aggregate boundaries and keeps invariants
// localised to each root.
//
// Aggregate roots in this codebase:
//
//   - [Plan]  — the business plan and all its wizard sections
//     (capital assets, salary roles, products, etc. are entities *within*
//     this aggregate and must not be referenced by other aggregates).
//   - [User]  — account identity, credentials, and access levels.
//
// # Sub-entity repositories
//
// The repository interfaces in internal/domain/repositories/ that cover wizard
// items (CapitalAssetRepository, SalaryRoleRepository, etc.) are NOT aggregate
// root repositories. They are convenience interfaces for the wizard's
// step-by-step persistence within the Plan aggregate boundary. The canonical
// aggregate-root repository is [repositories.PlanRepository].
package entities

import "github.com/google/uuid"

// AggregateRoot is a marker interface that every aggregate root must satisfy.
// Implementing this interface signals to readers that the type owns a
// consistency boundary and must be referenced by ID from other aggregates.
type AggregateRoot interface {
	// AggregateID returns the globally unique identity of this aggregate root.
	AggregateID() uuid.UUID
}
