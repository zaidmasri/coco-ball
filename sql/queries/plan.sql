-- name: SavePlan :exec
INSERT INTO plans (id, data, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at;

-- name: GetPlan :one
SELECT data FROM plans WHERE id = ? AND deleted_at IS NULL;

-- name: GetAllPlans :many
SELECT data FROM plans WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: GetUserPlans :many
SELECT p.data
FROM plans p
JOIN plan_access pa ON p.id = pa.plan_id
WHERE pa.user_id = ? AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;

-- name: SoftDeletePlan :execrows
UPDATE plans SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteCapitalAssetsByPlan :exec
UPDATE capital_assets SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteStartupCostsByPlan :exec
UPDATE startup_costs SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteFundingSourcesByPlan :exec
UPDATE funding_sources SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteSalaryRolesByPlan :exec
UPDATE salary_roles SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteBenefitsByPlan :exec
UPDATE benefits SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteProductsByPlan :exec
UPDATE products SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteInventoryPurchasesByPlan :exec
UPDATE inventory_purchases SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteDistributionsByPlan :exec
UPDATE distributions SET deleted_at = ? WHERE plan_id = ?;

-- name: SoftDeleteOperatingExpensesByPlan :exec
UPDATE operating_expenses SET deleted_at = ? WHERE plan_id = ?;

-- name: DeletePlanAccessByPlan :exec
DELETE FROM plan_access WHERE plan_id = ?;

-- name: DeleteStartingBalancesByPlan :exec
DELETE FROM starting_balances WHERE plan_id = ?;

-- name: DeletePayrollTaxRatesByPlan :exec
DELETE FROM payroll_tax_rates WHERE plan_id = ?;

-- name: DeleteSalesGrowthCurveByPlan :exec
DELETE FROM sales_growth_curve WHERE plan_id = ?;

-- name: DeleteWizardSectionsByPlan :exec
DELETE FROM wizard_sections WHERE plan_id = ?;
