-- name: CreateUser :exec
INSERT OR IGNORE INTO users (id, email, first_name, last_name, created_at) VALUES (?, ?, ?, ?, ?);

-- name: GetUserByID :one
SELECT email, first_name, last_name FROM users WHERE id = ? AND deleted_at IS NULL;

-- name: GetUserIDByEmail :one
SELECT id FROM users WHERE LOWER(email) = ? AND deleted_at IS NULL;

-- name: UpsertUserCredentials :exec
INSERT OR REPLACE INTO users_credentials (email, password_hash, created_at) VALUES (?, ?, ?);

-- name: GetUserCredentialsByEmail :one
SELECT u.id, uc.password_hash
FROM users u
JOIN users_credentials uc ON u.email = uc.email
WHERE LOWER(u.email) = ? AND u.deleted_at IS NULL;

-- name: UpdateUserName :exec
UPDATE users SET first_name = ?, last_name = ? WHERE id = ?;

-- name: DeleteSessionsByUser :exec
DELETE FROM sessions WHERE user_id = ?;

-- name: DeletePlanAccessByUser :exec
DELETE FROM plan_access WHERE user_id = ?;

-- name: DeleteUserCredentialsByEmail :exec
DELETE FROM users_credentials WHERE email = ?;

-- name: SoftDeleteUser :execrows
UPDATE users SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
