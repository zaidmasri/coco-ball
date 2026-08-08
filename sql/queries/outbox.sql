-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, aggregate_id, event_name, payload, occurred_at) VALUES (?, ?, ?, ?, ?);

-- name: GetUnpublishedOutboxEvents :many
SELECT id, aggregate_id, event_name, payload, occurred_at, published_at
FROM outbox_events
WHERE published_at IS NULL
ORDER BY occurred_at
LIMIT ?;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events SET published_at = ? WHERE id = ?;
