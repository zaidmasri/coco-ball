-- name: GetStartingBalancesRow :one
SELECT cash, accounts_receivable, prepaid_expenses, accounts_payable, accrued_expenses, current_step
FROM starting_balances WHERE plan_id = ?;

-- name: SaveStartingBalancesStep :exec
INSERT INTO starting_balances (plan_id, cash, accounts_receivable, prepaid_expenses, accounts_payable, accrued_expenses, current_step, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(plan_id) DO UPDATE SET
	cash = excluded.cash,
	accounts_receivable = excluded.accounts_receivable,
	prepaid_expenses = excluded.prepaid_expenses,
	accounts_payable = excluded.accounts_payable,
	accrued_expenses = excluded.accrued_expenses,
	current_step = excluded.current_step,
	updated_at = excluded.updated_at;
