// Package events defines the domain event types for all aggregate roots.
// Only aggregate roots (Plan, User) may emit events; entities within an
// aggregate boundary (PlanInvite, Session, etc.) must not carry event buffers.
package events

import (
	"time"

	"uuid"
)

// DomainEvent is implemented by every domain event type. Repositories must
// drain an aggregate's events via PullEvents() and persist each one to the
// outbox table in the same transaction that saves the aggregate.
type DomainEvent interface {
	EventName() string
	AggregateID() uuid.UUID
	OccurredAt() time.Time
}

// BaseEvent carries the fields common to all domain events. Embed it in
// every concrete event struct.
type BaseEvent struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

// NewBaseEvent constructs a BaseEvent stamped with the current time.
func NewBaseEvent(aggregateID uuid.UUID) BaseEvent {
	return BaseEvent{
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
	}
}

func (e BaseEvent) AggregateID() uuid.UUID { return e.aggregateID }
func (e BaseEvent) OccurredAt() time.Time  { return e.occurredAt }
