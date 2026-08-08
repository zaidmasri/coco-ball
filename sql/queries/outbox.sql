-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, event_name, payload, created_at) VALUES (?, ?, ?, ?);
