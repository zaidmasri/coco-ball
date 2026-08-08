-- name: MaxSalaryRoleSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM salary_roles WHERE plan_id = ?;

-- name: CreateSalaryRoleDraft :exec
INSERT INTO salary_roles (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindSalaryRoleDraft :one
SELECT id, role, is_contractor, headcount, monthly_pay, growth_yr2, growth_yr3, status, current_step
FROM salary_roles
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetSalaryRole :one
SELECT id, role, is_contractor, headcount, monthly_pay, growth_yr2, growth_yr3, status, current_step
FROM salary_roles
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveSalaryRoleStep :execrows
UPDATE salary_roles
SET role = ?, is_contractor = ?, headcount = ?, monthly_pay = ?, growth_yr2 = ?, growth_yr3 = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteSalaryRoles :many
SELECT id, role, is_contractor, headcount, monthly_pay, growth_yr2, growth_yr3, status, current_step
FROM salary_roles
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteSalaryRole :execrows
UPDATE salary_roles SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
