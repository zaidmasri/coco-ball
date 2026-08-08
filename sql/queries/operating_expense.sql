-- name: MaxOperatingExpenseSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM operating_expenses WHERE plan_id = ?;

-- name: CreateOperatingExpenseDraft :exec
INSERT INTO operating_expenses (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindOperatingExpenseDraft :one
SELECT id, name, monthly_amount, growth_type, annual_rate, status, current_step
FROM operating_expenses
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetOperatingExpense :one
SELECT id, name, monthly_amount, growth_type, annual_rate, status, current_step
FROM operating_expenses
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveOperatingExpenseStep :execrows
UPDATE operating_expenses
SET name = ?, monthly_amount = ?, growth_type = ?, annual_rate = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteOperatingExpenses :many
SELECT id, name, monthly_amount, growth_type, annual_rate, status, current_step
FROM operating_expenses
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteOperatingExpense :execrows
UPDATE operating_expenses SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
