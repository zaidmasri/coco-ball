package repositories

// OutboxEvent is a row from the outbox_events table. Payload is opaque JSON
// — the domain layer doesn't parse it; the interface-layer worker that
// consumes it does.
type OutboxEvent struct {
	ID          string
	AggregateID string
	EventName   string
	Payload     string
	OccurredAt  int64
}

// OutboxRepository reads and marks the transactional outbox written by
// aggregate repositories (see infrastructure/sqlite/outbox.go's
// insertOutboxEvents). Implementations must not filter or reorder beyond
// "unpublished, oldest first" — deciding what to do with an event is the
// consumer's job, not the repository's.
type OutboxRepository interface {
	GetUnpublished(limit int) ([]OutboxEvent, error)
	MarkPublished(id string) error
}
