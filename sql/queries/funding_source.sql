-- name: MaxFundingSourceSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM funding_sources WHERE plan_id = ?;

-- name: CreateFundingSourceDraft :exec
INSERT INTO funding_sources (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindFundingSourceDraft :one
SELECT id, name, amount, interest_rate, term_months, status, current_step
FROM funding_sources
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetFundingSource :one
SELECT id, name, amount, interest_rate, term_months, status, current_step
FROM funding_sources
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveFundingSourceStep :execrows
UPDATE funding_sources
SET name = ?, amount = ?, interest_rate = ?, term_months = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteFundingSources :many
SELECT id, name, amount, interest_rate, term_months, status, current_step
FROM funding_sources
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteFundingSource :execrows
UPDATE funding_sources SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
