package queries

import "github.com/google/uuid"

// GetPlan retrieves a single plan by ID.
type GetPlan struct {
	PlanID uuid.UUID
}

// GetUserPlans retrieves all plans visible to a given user.
type GetUserPlans struct {
	UserID uuid.UUID
}

// GetAllPlans retrieves every plan (admin use).
type GetAllPlans struct{}
