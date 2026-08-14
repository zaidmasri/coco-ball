package commands

import "github.com/google/uuid"

// DeletePlan is the command to delete a plan and all its data.
type DeletePlan struct {
	PlanID uuid.UUID
}

// DeletePlanResult reports whether the deletion succeeded.
type DeletePlanResult struct {
	Success bool
}
