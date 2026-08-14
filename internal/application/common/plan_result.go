// Package common holds output DTOs shared between the command and query
// sides of the application layer. A Result is not the domain entity — it's
// a display-oriented shape safe to hand to callers that must not be able to
// invoke domain behavior (mutation methods, PullEvents) on what they hold.
package common

import "github.com/google/uuid"

// PlanResult is the write-side acknowledgement shape for plan commands. It
// deliberately does not carry the full aggregate (wizard sections, domain
// events, mutation methods) — callers of CreatePlan/UpdatePlan only need
// enough to redirect and grant access. Read paths that need the full
// aggregate (report pages, wizard controllers) go through
// PlanService.Get/GetAll/GetUserPlans instead, which return *domain.Plan
// directly — see AGENTS.md's "Application Layer: go-ddd Comparison" section
// for why the read side does not use this type.
type PlanResult struct {
	ID            uuid.UUID
	Name          string
	StartingMonth int
	StartingYear  int
	OwnerID       uuid.UUID
}
