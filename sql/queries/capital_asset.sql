-- name: MaxCapitalAssetSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM capital_assets WHERE plan_id = ?;

-- name: CreateCapitalAssetDraft :exec
INSERT INTO capital_assets (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindCapitalAssetDraft :one
SELECT id, name, purchase_cost, useful_life_months, salvage_value, purchase_month_index, depreciation_method, status, current_step
FROM capital_assets
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetCapitalAsset :one
SELECT id, name, purchase_cost, useful_life_months, salvage_value, purchase_month_index, depreciation_method, status, current_step
FROM capital_assets
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveCapitalAssetStep :execrows
UPDATE capital_assets
SET name = ?, purchase_cost = ?, useful_life_months = ?, salvage_value = ?, purchase_month_index = ?, depreciation_method = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteCapitalAssets :many
SELECT id, name, purchase_cost, useful_life_months, salvage_value, purchase_month_index, depreciation_method, status, current_step
FROM capital_assets
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteCapitalAsset :execrows
UPDATE capital_assets SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
