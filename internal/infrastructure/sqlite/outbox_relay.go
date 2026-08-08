package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

const outboxRelayBatchSize = 100

// OutboxRelay polls outbox_events for unpublished rows and publishes them.
// The publish step is currently a log line; swap publishEvent for a real
// message bus/webhook call once one exists.
type OutboxRelay struct {
	queries  *db.Queries
	interval time.Duration
}

// NewOutboxRelay constructs a relay that polls the outbox on the given
// interval using a connection shared with the rest of the sqlite package.
func NewOutboxRelay(conn *sql.DB, interval time.Duration) *OutboxRelay {
	return &OutboxRelay{queries: db.New(conn), interval: interval}
}

// Start runs the poll loop in a background goroutine until ctx is canceled.
func (r *OutboxRelay) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.publishBatch(ctx); err != nil {
					log.Printf("outbox relay: %v", err)
				}
			}
		}
	}()
}

// publishBatch drains up to outboxRelayBatchSize unpublished events, marking
// each published as it succeeds so a crash mid-batch only reprocesses the
// remainder (at-least-once delivery).
func (r *OutboxRelay) publishBatch(ctx context.Context) error {
	events, err := r.queries.GetUnpublishedOutboxEvents(ctx, outboxRelayBatchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch unpublished outbox events: %w", err)
	}

	for _, evt := range events {
		if err := r.publishEvent(evt); err != nil {
			return fmt.Errorf("failed to publish outbox event %s: %w", evt.ID, err)
		}

		if err := r.queries.MarkOutboxEventPublished(ctx, db.MarkOutboxEventPublishedParams{
			PublishedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
			ID:          evt.ID,
		}); err != nil {
			return fmt.Errorf("failed to mark outbox event %s published: %w", evt.ID, err)
		}
	}

	return nil
}

// publishEvent is a stand-in for a real publish step (message bus, webhook,
// etc.). It never errors today; the error return exists so callers don't
// need to change once a real transport is wired in.
func (r *OutboxRelay) publishEvent(evt db.OutboxEvent) error {
	log.Printf("outbox: publishing event %s %q aggregate=%s", evt.ID, evt.EventName, evt.AggregateID)
	return nil
}
