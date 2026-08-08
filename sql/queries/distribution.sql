-- name: MaxDistributionSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM distributions WHERE plan_id = ?;

-- name: CreateDistributionDraft :exec
INSERT INTO distributions (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindDistributionDraft :one
SELECT id, name, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM distributions
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetDistribution :one
SELECT id, name, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM distributions
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveDistributionStep :execrows
UPDATE distributions
SET name = ?, monthly_amount = ?, growth_yr2 = ?, growth_yr3 = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteDistributions :many
SELECT id, name, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM distributions
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteDistribution :execrows
UPDATE distributions SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
