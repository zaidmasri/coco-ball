package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// Migration is a single schema or data change. Exactly one of SQL or Fn is
// set. SQL migrations are run via a single db.Exec call; Fn migrations run
// arbitrary Go code (used for data backfills that raw SQL can't express,
// e.g. reading the plans.data JSON blob).
type Migration struct {
	SQL string
	Fn  func(db *sql.DB) error
}

var migrations = []Migration{
	// Migration 1: Create migrations tracking table (MUST BE FIRST)
	{SQL: `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`},

	// Migration 2: Create users table
	{SQL: `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		created_at INTEGER NOT NULL
	)`},

	// Migration 3: Create users_credentials table
	{SQL: `CREATE TABLE IF NOT EXISTS users_credentials (
		email TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE
	)`},

	// Migration 4: Create sessions table
	{SQL: `CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`},

	// Migration 5: Create plans table
	{SQL: `CREATE TABLE IF NOT EXISTS plans (
		id TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`},

	// Migration 6: Create plan_access table
	{SQL: `CREATE TABLE IF NOT EXISTS plan_access (
		plan_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		access_level TEXT NOT NULL,
		invited_at INTEGER NOT NULL,
		PRIMARY KEY (plan_id, user_id),
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`},

	// Migration 7: Add first name to users
	{SQL: `ALTER TABLE users ADD COLUMN first_name TEXT NOT NULL DEFAULT ''`},

	// Migration 8: Add last name to users
	{SQL: `ALTER TABLE users ADD COLUMN last_name TEXT NOT NULL DEFAULT ''`},

	// Migration 9: Create plan_invites table
	{SQL: `CREATE TABLE IF NOT EXISTS plan_invites (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		email TEXT NOT NULL,
		access_level TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		invited_by TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		responded_at INTEGER,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE,
		FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
	)`},

	// Migration 10: Index invites by email for onboarding lookups
	{SQL: `CREATE INDEX IF NOT EXISTS idx_plan_invites_email ON plan_invites(email)`},

	// Migration 11: Create capital_assets table (normalizes Starting
	// Point's "Fixed Assets" section out of the plans.data JSON blob so
	// the new wizard can save one item at a time without racing other
	// sections).
	{SQL: `CREATE TABLE IF NOT EXISTS capital_assets (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		purchase_cost INTEGER NOT NULL DEFAULT 0,
		useful_life_months INTEGER NOT NULL DEFAULT 0,
		salvage_value INTEGER NOT NULL DEFAULT 0,
		purchase_month_index INTEGER NOT NULL DEFAULT 0,
		depreciation_method TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 12: Index capital_assets by plan for list/draft lookups
	{SQL: `CREATE INDEX IF NOT EXISTS idx_capital_assets_plan_id ON capital_assets(plan_id)`},

	// Migration 13: Create startup_costs table (Starting Point's
	// "Operating Capital" section)
	{SQL: `CREATE TABLE IF NOT EXISTS startup_costs (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		amount INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 14: Index startup_costs by plan
	{SQL: `CREATE INDEX IF NOT EXISTS idx_startup_costs_plan_id ON startup_costs(plan_id)`},

	// Migration 15: Create funding_sources table
	{SQL: `CREATE TABLE IF NOT EXISTS funding_sources (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		amount INTEGER NOT NULL DEFAULT 0,
		interest_rate REAL NOT NULL DEFAULT 0,
		term_months INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 16: Index funding_sources by plan
	{SQL: `CREATE INDEX IF NOT EXISTS idx_funding_sources_plan_id ON funding_sources(plan_id)`},

	// Migration 17: Create starting_balances table (Starting Point's
	// "Cash on Hand" section - a singleton per plan, keyed by plan_id
	// rather than its own id since there's exactly one row per plan)
	{SQL: `CREATE TABLE IF NOT EXISTS starting_balances (
		plan_id TEXT PRIMARY KEY,
		cash INTEGER NOT NULL DEFAULT 0,
		accounts_receivable INTEGER NOT NULL DEFAULT 0,
		prepaid_expenses INTEGER NOT NULL DEFAULT 0,
		accounts_payable INTEGER NOT NULL DEFAULT 0,
		accrued_expenses INTEGER NOT NULL DEFAULT 0,
		current_step INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 18: Create starting_point_sections table - the single
	// source of truth for each section's explicit "I'm done" marker,
	// independent of how many items it currently has (deleting a
	// section's last item does not un-complete it).
	{SQL: `CREATE TABLE IF NOT EXISTS starting_point_sections (
		plan_id TEXT NOT NULL,
		section TEXT NOT NULL,
		completed_at INTEGER,
		PRIMARY KEY (plan_id, section),
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 19: Backfill capital_assets/startup_costs/funding_sources/
	// starting_balances/starting_point_sections from every existing
	// plan's plans.data JSON blob, since migrations 11-18 introduced new
	// tables that must not orphan pre-existing Starting Point data.
	{Fn: backfillStartingPointData},
}

// GetMigrations returns all migrations in application order.
func GetMigrations() []Migration {
	return migrations
}

// legacyPlanStartingPoint mirrors the subset of domain.Plan's old planJSON
// fields relevant to Starting Point, for decoding the pre-normalization
// JSON blob during backfill. Field names match planJSON's json tags
// exactly; nested struct field names have no json tags in the domain
// package, so they marshal/unmarshal using their literal Go names.
type legacyPlanStartingPoint struct {
	FuturePurchases []struct {
		Name               string
		PurchaseCost       int64
		UsefulLifeMonths   int
		SalvageValue       int64
		PurchaseMonthIndex int64
		DepreciationMethod string
	} `json:"futurePurchases"`
	StartupCosts []struct {
		Name   string
		Amount int64
	} `json:"startupCosts"`
	FundingSources []struct {
		Name         string
		Amount       int64
		InterestRate float64
		TermMonths   int
	} `json:"fundingSources"`
	StartingBalances struct {
		Cash               int64
		AccountsReceivable int64
		PrepaidExpenses    int64
		AccountsPayable    int64
		AccruedExpenses    int64
	} `json:"startingBalances"`
}

// backfillStartingPointData populates the new normalized tables from every
// existing plan's JSON blob so pre-existing Starting Point data isn't
// orphaned by migrations 11-18. Plans with zero items in all three
// repeatable sections and an all-zero starting balance are left with no
// rows at all (and therefore show as "Not started" on the new summary
// page) - this is indistinguishable from a plan that never visited
// Starting Point, which is a deliberate, low-blast-radius tradeoff (one
// re-click) rather than guessing "Complete".
func backfillStartingPointData(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("backfill: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, data, updated_at FROM plans")
	if err != nil {
		return fmt.Errorf("backfill: failed to query plans: %w", err)
	}

	type planRow struct {
		id        string
		data      []byte
		updatedAt int64
	}
	var plans []planRow
	for rows.Next() {
		var pr planRow
		if err := rows.Scan(&pr.id, &pr.data, &pr.updatedAt); err != nil {
			rows.Close()
			return fmt.Errorf("backfill: failed to scan plan row: %w", err)
		}
		plans = append(plans, pr)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("backfill: failed to iterate plan rows: %w", err)
	}
	rows.Close()

	for _, pr := range plans {
		var legacy legacyPlanStartingPoint
		if err := json.Unmarshal(pr.data, &legacy); err != nil {
			return fmt.Errorf("backfill: failed to decode plan %s: %w", pr.id, err)
		}

		now := pr.updatedAt

		if len(legacy.FuturePurchases) > 0 {
			for i, a := range legacy.FuturePurchases {
				if _, err := tx.Exec(
					`INSERT INTO capital_assets (id, plan_id, name, purchase_cost, useful_life_months, salvage_value, purchase_month_index, depreciation_method, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, a.Name, a.PurchaseCost, a.UsefulLifeMonths, a.SalvageValue, a.PurchaseMonthIndex, a.DepreciationMethod, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert capital_asset for plan %s: %w", pr.id, err)
				}
			}
			if err := markSectionComplete(tx, pr.id, domain.SectionFixedAssets, now); err != nil {
				return err
			}
		}

		if len(legacy.StartupCosts) > 0 {
			for i, c := range legacy.StartupCosts {
				if _, err := tx.Exec(
					`INSERT INTO startup_costs (id, plan_id, name, amount, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, c.Name, c.Amount, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert startup_cost for plan %s: %w", pr.id, err)
				}
			}
			if err := markSectionComplete(tx, pr.id, domain.SectionStartupCosts, now); err != nil {
				return err
			}
		}

		if len(legacy.FundingSources) > 0 {
			for i, f := range legacy.FundingSources {
				if _, err := tx.Exec(
					`INSERT INTO funding_sources (id, plan_id, name, amount, interest_rate, term_months, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, f.Name, f.Amount, f.InterestRate, f.TermMonths, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert funding_source for plan %s: %w", pr.id, err)
				}
			}
			if err := markSectionComplete(tx, pr.id, domain.SectionFundingSources, now); err != nil {
				return err
			}
		}

		sb := legacy.StartingBalances
		if sb.Cash != 0 || sb.AccountsReceivable != 0 || sb.PrepaidExpenses != 0 || sb.AccountsPayable != 0 || sb.AccruedExpenses != 0 {
			if _, err := tx.Exec(
				`INSERT INTO starting_balances (plan_id, cash, accounts_receivable, prepaid_expenses, accounts_payable, accrued_expenses, current_step, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`,
				pr.id, sb.Cash, sb.AccountsReceivable, sb.PrepaidExpenses, sb.AccountsPayable, sb.AccruedExpenses, now, now,
			); err != nil {
				return fmt.Errorf("backfill: failed to insert starting_balances for plan %s: %w", pr.id, err)
			}
			if err := markSectionComplete(tx, pr.id, domain.SectionCashOnHand, now); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func markSectionComplete(tx *sql.Tx, planID, section string, completedAt int64) error {
	_, err := tx.Exec(
		`INSERT INTO starting_point_sections (plan_id, section, completed_at) VALUES (?, ?, ?)
		 ON CONFLICT(plan_id, section) DO UPDATE SET completed_at = excluded.completed_at`,
		planID, section, completedAt,
	)
	if err != nil {
		return fmt.Errorf("backfill: failed to mark section %s complete for plan %s: %w", section, planID, err)
	}
	return nil
}
