package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
)

// CreatePlan is the command to create a new business plan. PlanService.
// CreatePlan verifies OwnerID names a real User (producing a
// domain.VerifiedUser) and turns the rest into a validated domain.Plan via
// domain.NewPlan — callers must not construct the entity themselves.
type CreatePlan struct {
	Name          string
	StartingMonth int
	StartingYear  int
	OwnerID       uuid.UUID
}

// CreatePlanResult wraps the created plan's Result. The controller uses
// Result.ID for the redirect and for granting the creator Owner access.
type CreatePlanResult struct {
	Result *common.PlanResult
}
