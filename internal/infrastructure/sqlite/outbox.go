package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// insertOutboxEvents stores domain events in the outbox table. Call it with
// queries bound to the same transaction as the aggregate write so the state
// change and its events commit or roll back together. now is the same
// timestamp used for the aggregate's own created_at/updated_at, keeping the
// outbox row's created_at consistent within one transaction rather than
// using each event's own OccurredAt().
func insertOutboxEvents(ctx context.Context, q *db.Queries, events []domevents.DomainEvent, now int64) error {
	for _, evt := range events {
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate outbox event id: %w", err)
		}

		if err := q.InsertOutboxEvent(ctx, db.InsertOutboxEventParams{
			ID:        id.String(),
			EventName: evt.EventName(),
			Payload:   string(payload),
			CreatedAt: now,
		}); err != nil {
			return fmt.Errorf("failed to insert outbox event %q: %w", evt.EventName(), err)
		}
	}
	return nil
}
