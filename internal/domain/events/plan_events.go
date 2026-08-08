package events

import "github.com/google/uuid"

const (
	PlanCreatedEventName = "plan.created"
	PlanUpdatedEventName = "plan.updated"
	PlanDeletedEventName = "plan.deleted"
)

// PlanCreated is emitted the first time a plan is persisted.
type PlanCreated struct {
	BaseEvent
	Name    string
	OwnerID uuid.UUID
}

func NewPlanCreated(planID uuid.UUID, name string, ownerID uuid.UUID) PlanCreated {
	return PlanCreated{
		BaseEvent: NewBaseEvent(planID),
		Name:      name,
		OwnerID:   ownerID,
	}
}

func (e PlanCreated) EventName() string { return PlanCreatedEventName }

// PlanUpdated is emitted when a plan's core details change.
type PlanUpdated struct {
	BaseEvent
}

func NewPlanUpdated(planID uuid.UUID) PlanUpdated {
	return PlanUpdated{BaseEvent: NewBaseEvent(planID)}
}

func (e PlanUpdated) EventName() string { return PlanUpdatedEventName }

// PlanDeleted is emitted when a plan is soft-deleted.
type PlanDeleted struct {
	BaseEvent
}

func NewPlanDeleted(planID uuid.UUID) PlanDeleted {
	return PlanDeleted{BaseEvent: NewBaseEvent(planID)}
}

func (e PlanDeleted) EventName() string { return PlanDeletedEventName }
