-- name: MaxStartupCostSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM startup_costs WHERE plan_id = ?;

-- name: CreateStartupCostDraft :exec
INSERT INTO startup_costs (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindStartupCostDraft :one
SELECT id, name, amount, status, current_step
FROM startup_costs
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetStartupCost :one
SELECT id, name, amount, status, current_step
FROM startup_costs
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveStartupCostStep :execrows
UPDATE startup_costs
SET name = ?, amount = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteStartupCosts :many
SELECT id, name, amount, status, current_step
FROM startup_costs
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteStartupCost :execrows
UPDATE startup_costs SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
