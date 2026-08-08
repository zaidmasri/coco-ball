package commands

import "github.com/google/uuid"

// CreatePlan is the command to create a new business plan.
type CreatePlan struct {
	Name          string
	StartingMonth int
	StartingYear  int
	OwnerID       uuid.UUID
}

// UpdatePlan is the command to change a plan's core details.
type UpdatePlan struct {
	PlanID        uuid.UUID
	Name          string
	StartingMonth int
	StartingYear  int
}

// DeletePlan is the command to soft-delete a plan and all its data.
type DeletePlan struct {
	PlanID uuid.UUID
}
