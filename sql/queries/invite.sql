-- name: CreateInvite :exec
INSERT INTO plan_invites (id, plan_id, email, access_level, status, invited_by, created_at, responded_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULL);

-- name: GetInvite :one
SELECT plan_id, email, access_level, status, invited_by, created_at, responded_at
FROM plan_invites WHERE id = ?;

-- name: GetInvitesForPlan :many
SELECT id, email, access_level, status, invited_by, created_at, responded_at
FROM plan_invites WHERE plan_id = ? ORDER BY created_at DESC;

-- name: GetPendingInvitesForEmail :many
SELECT id, plan_id, access_level, status, invited_by, created_at, responded_at
FROM plan_invites WHERE LOWER(email) = ? AND status = ? ORDER BY created_at DESC;

-- name: UpdateInviteStatus :execrows
UPDATE plan_invites SET status = ?, responded_at = ? WHERE id = ?;
