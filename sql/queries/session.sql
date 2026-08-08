-- name: SaveSession :exec
INSERT OR REPLACE INTO sessions (id, user_id, created_at, expires_at) VALUES (?, ?, ?, ?);

-- name: GetSession :one
SELECT user_id, created_at, expires_at FROM sessions WHERE id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;
