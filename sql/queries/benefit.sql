-- name: MaxBenefitSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) + 1 FROM benefits WHERE plan_id = ?;

-- name: CreateBenefitDraft :exec
INSERT INTO benefits (id, plan_id, status, current_step, sort_order, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?, ?);

-- name: FindBenefitDraft :one
SELECT id, type, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM benefits
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetBenefit :one
SELECT id, type, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM benefits
WHERE id = ? AND deleted_at IS NULL;

-- name: SaveBenefitStep :execrows
UPDATE benefits
SET type = ?, monthly_amount = ?, growth_yr2 = ?, growth_yr3 = ?, status = ?, current_step = ?, updated_at = ?
WHERE id = ?;

-- name: ListCompleteBenefits :many
SELECT id, type, monthly_amount, growth_yr2, growth_yr3, status, current_step
FROM benefits
WHERE plan_id = ? AND status = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: DeleteBenefit :execrows
UPDATE benefits SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
