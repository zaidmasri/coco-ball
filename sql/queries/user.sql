-- name: CreateUser :exec
INSERT OR IGNORE INTO users (id, email, first_name, last_name, created_at) VALUES (?, ?, ?, ?, ?);

-- name: GetUserByID :one
SELECT email, first_name, last_name FROM users WHERE id = ?;

-- name: GetUserIDByEmail :one
SELECT id FROM users WHERE LOWER(email) = ?;

-- name: UpsertUserCredentials :exec
INSERT OR REPLACE INTO users_credentials (email, password_hash, created_at) VALUES (?, ?, ?);

-- name: GetUserCredentialsByEmail :one
SELECT u.id, uc.password_hash
FROM users u
JOIN users_credentials uc ON u.email = uc.email
WHERE LOWER(u.email) = ?;
