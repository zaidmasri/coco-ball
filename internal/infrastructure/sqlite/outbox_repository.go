package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// OutboxRepository implements repositories.OutboxRepository.
type OutboxRepository struct {
	queries *db.Queries
}

func NewOutboxRepository(conn *sql.DB) repositories.OutboxRepository {
	return &OutboxRepository{queries: db.New(conn)}
}

var _ repositories.OutboxRepository = (*OutboxRepository)(nil)

func (r *OutboxRepository) GetUnpublished(limit int) ([]repositories.OutboxEvent, error) {
	rows, err := r.queries.GetUnpublishedOutboxEvents(context.Background(), int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get unpublished outbox events: %w", err)
	}

	events := make([]repositories.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, repositories.OutboxEvent{
			ID:          row.ID,
			AggregateID: row.AggregateID,
			EventName:   row.EventName,
			Payload:     row.Payload,
			OccurredAt:  row.OccurredAt,
		})
	}
	return events, nil
}

func (r *OutboxRepository) MarkPublished(id string) error {
	if err := r.queries.MarkOutboxEventPublished(context.Background(), db.MarkOutboxEventPublishedParams{
		PublishedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:          id,
	}); err != nil {
		return fmt.Errorf("failed to mark outbox event %s published: %w", id, err)
	}
	return nil
}
