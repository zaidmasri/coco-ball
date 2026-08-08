-- name: GetSalesGrowthCurveRow :one
SELECT y1_q1, y1_q2, y1_q3, y1_q4, growth_yr2, growth_yr3, current_step
FROM sales_growth_curve WHERE plan_id = ?;

-- name: SaveSalesGrowthCurveStep :exec
INSERT INTO sales_growth_curve (plan_id, y1_q1, y1_q2, y1_q3, y1_q4, growth_yr2, growth_yr3, current_step, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(plan_id) DO UPDATE SET
	y1_q1 = excluded.y1_q1,
	y1_q2 = excluded.y1_q2,
	y1_q3 = excluded.y1_q3,
	y1_q4 = excluded.y1_q4,
	growth_yr2 = excluded.growth_yr2,
	growth_yr3 = excluded.growth_yr3,
	current_step = excluded.current_step,
	updated_at = excluded.updated_at;
