-- name: CountPlanExists :one
SELECT COUNT(*) FROM plans WHERE id = ? AND deleted_at IS NULL;

-- name: CountUserExists :one
SELECT COUNT(*) FROM users WHERE id = ?;

-- name: GrantAccess :exec
INSERT OR REPLACE INTO plan_access (plan_id, user_id, access_level, invited_at) VALUES (?, ?, ?, ?);

-- name: GetAccess :one
SELECT access_level, invited_at FROM plan_access WHERE plan_id = ? AND user_id = ?;

-- name: GetPlanAccess :many
SELECT user_id, access_level, invited_at FROM plan_access WHERE plan_id = ?;

-- name: GetPlansForUser :many
SELECT p.data
FROM plans p
JOIN plan_access pa ON p.id = pa.plan_id
WHERE pa.user_id = ? AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;
