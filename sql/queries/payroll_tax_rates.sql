-- name: GetPayrollTaxRatesRow :one
SELECT social_security_rate, medicare_rate, futa_rate, suta_rate, current_step
FROM payroll_tax_rates WHERE plan_id = ?;

-- name: SavePayrollTaxRatesStep :exec
INSERT INTO payroll_tax_rates (plan_id, social_security_rate, medicare_rate, futa_rate, suta_rate, current_step, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(plan_id) DO UPDATE SET
	social_security_rate = excluded.social_security_rate,
	medicare_rate = excluded.medicare_rate,
	futa_rate = excluded.futa_rate,
	suta_rate = excluded.suta_rate,
	current_step = excluded.current_step,
	updated_at = excluded.updated_at;
