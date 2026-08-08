-- name: MaxInventoryPurchaseSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM inventory_purchases WHERE plan_id = ?;

-- name: CreateInventoryPurchaseDraft :exec
INSERT INTO inventory_purchases (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindInventoryPurchaseDraft :one
SELECT id, category, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM inventory_purchases
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetInventoryPurchase :one
SELECT id, category, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM inventory_purchases
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveInventoryPurchaseStep :execrows
UPDATE inventory_purchases
SET category = ?, monthly_amount = ?, growth_yr2 = ?, growth_yr3 = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteInventoryPurchases :many
SELECT id, category, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM inventory_purchases
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteInventoryPurchase :execrows
UPDATE inventory_purchases SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
