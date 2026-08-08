-- name: MaxProductSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM products WHERE plan_id = ?;

-- name: CreateProductDraft :exec
INSERT INTO products (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindProductDraft :one
SELECT id, name, month1_units, price_per_unit, cost_per_unit, status, current_step
FROM products
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetProduct :one
SELECT id, name, month1_units, price_per_unit, cost_per_unit, status, current_step
FROM products
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveProductStep :execrows
UPDATE products
SET name = ?, month1_units = ?, price_per_unit = ?, cost_per_unit = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteProducts :many
SELECT id, name, month1_units, price_per_unit, cost_per_unit, status, current_step
FROM products
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteProduct :execrows
UPDATE products SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
