-- name: GrantAccess :exec
INSERT OR REPLACE INTO plan_access (plan_id, user_id, access_level, invited_at) VALUES (?, ?, ?, ?);

-- name: GetAccess :one
SELECT access_level, invited_at FROM plan_access WHERE plan_id = ? AND user_id = ?;

-- name: GetPlanAccess :many
SELECT user_id, access_level, invited_at FROM plan_access WHERE plan_id = ?;
