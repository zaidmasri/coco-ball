package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
)

// UpdatePlan is the command to change a plan's core details.
// PlanService.UpdatePlan loads the aggregate and calls
// Plan.ChangeCoreDetails — callers must not mutate the entity themselves.
type UpdatePlan struct {
	PlanID        uuid.UUID
	Name          string
	StartingMonth int
	StartingYear  int
}

// UpdatePlanResult wraps the updated plan's Result.
type UpdatePlanResult struct {
	Result *common.PlanResult
}
