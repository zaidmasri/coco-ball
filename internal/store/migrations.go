package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
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

	// Migrations 20-31: normalize Payroll/Sales Forecast/Cash Flow out of
	// the plans.data JSON blob, the same way migrations 11-18 normalized
	// Starting Point - one table per repeatable entity (with
	// status/current_step/sort_order for wizard draft/resume support) and
	// one singleton table per per-plan entity (Payroll Tax Rates, Sales
	// Growth Curve).
	{SQL: `CREATE TABLE IF NOT EXISTS salary_roles (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT '',
		is_contractor INTEGER NOT NULL DEFAULT 0,
		headcount INTEGER NOT NULL DEFAULT 0,
		monthly_pay INTEGER NOT NULL DEFAULT 0,
		growth_yr2 REAL NOT NULL DEFAULT 0,
		growth_yr3 REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_salary_roles_plan_id ON salary_roles(plan_id)`},

	{SQL: `CREATE TABLE IF NOT EXISTS benefits (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT '',
		monthly_amount INTEGER NOT NULL DEFAULT 0,
		growth_yr2 REAL NOT NULL DEFAULT 0,
		growth_yr3 REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_benefits_plan_id ON benefits(plan_id)`},

	// Payroll Tax Rates is a singleton per plan (like starting_balances).
	{SQL: `CREATE TABLE IF NOT EXISTS payroll_tax_rates (
		plan_id TEXT PRIMARY KEY,
		social_security_rate REAL NOT NULL DEFAULT 0,
		medicare_rate REAL NOT NULL DEFAULT 0,
		futa_rate REAL NOT NULL DEFAULT 0,
		suta_rate REAL NOT NULL DEFAULT 0,
		current_step INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	{SQL: `CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		month1_units INTEGER NOT NULL DEFAULT 0,
		price_per_unit INTEGER NOT NULL DEFAULT 0,
		cost_per_unit INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_products_plan_id ON products(plan_id)`},

	// Sales Growth Curve is a singleton per plan. Every "growth after Year
	// 1" schedule app-wide (this one included) is standardized to a fixed
	// Year 2 + Year 3, so this stores fixed columns rather than a JSON
	// array - Sales Forecast's old "Add Year" button (Year 4+) is removed.
	{SQL: `CREATE TABLE IF NOT EXISTS sales_growth_curve (
		plan_id TEXT PRIMARY KEY,
		y1_q1 REAL NOT NULL DEFAULT 0,
		y1_q2 REAL NOT NULL DEFAULT 0,
		y1_q3 REAL NOT NULL DEFAULT 0,
		y1_q4 REAL NOT NULL DEFAULT 0,
		growth_yr2 REAL NOT NULL DEFAULT 0,
		growth_yr3 REAL NOT NULL DEFAULT 0,
		current_step INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	{SQL: `CREATE TABLE IF NOT EXISTS inventory_purchases (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		category TEXT NOT NULL DEFAULT '',
		monthly_amount INTEGER NOT NULL DEFAULT 0,
		growth_yr2 REAL NOT NULL DEFAULT 0,
		growth_yr3 REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_inventory_purchases_plan_id ON inventory_purchases(plan_id)`},

	{SQL: `CREATE TABLE IF NOT EXISTS distributions (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		monthly_amount INTEGER NOT NULL DEFAULT 0,
		growth_yr2 REAL NOT NULL DEFAULT 0,
		growth_yr3 REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_distributions_plan_id ON distributions(plan_id)`},

	// Migration 32: wizard_sections generalizes starting_point_sections to
	// cover all four wizard hubs (Starting Point, Payroll, Sales Forecast,
	// Cash Flow), each identified by a "hub" key, instead of one
	// hub-specific completion table per hub.
	{SQL: `CREATE TABLE IF NOT EXISTS wizard_sections (
		plan_id TEXT NOT NULL,
		hub TEXT NOT NULL,
		section TEXT NOT NULL,
		completed_at INTEGER,
		PRIMARY KEY (plan_id, hub, section),
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},

	// Migration 33: move starting_point_sections' rows into wizard_sections
	// (hub='starting-point') and drop the now-redundant table.
	{Fn: migrateStartingPointSectionsToWizardSections},

	// Migration 34: backfill salary_roles/benefits/payroll_tax_rates/
	// products/sales_growth_curve/inventory_purchases/distributions (and
	// their wizard_sections completion rows) from every existing plan's
	// plans.data JSON blob, since migrations 20-32 introduced new tables
	// that must not orphan pre-existing Payroll/Sales Forecast/Cash Flow
	// data. Every AnnualGrowth-shaped field is truncated to its first two
	// entries (Year 2, Year 3) to match the new fixed growth_yr2/growth_yr3
	// columns.
	{Fn: backfillPayrollSalesForecastCashFlowData},

	// Migration 35: normalize Operating Expenses out of the plans.data
	// JSON blob, the same way migrations 11-18/20-32 normalized Starting
	// Point/Payroll/Sales Forecast/Cash Flow. Operating Expenses uses a
	// single annual growth rate (GrowthStrategy: Flat or AnnualStepPercent)
	// rather than the fixed Year2/Year3 pair the other hubs' entities use,
	// so this table stores growth_type/annual_rate instead of
	// growth_yr2/growth_yr3.
	{SQL: `CREATE TABLE IF NOT EXISTS operating_expenses (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		monthly_amount INTEGER NOT NULL DEFAULT 0,
		growth_type TEXT NOT NULL DEFAULT 'Flat',
		annual_rate REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		current_step INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_operating_expenses_plan_id ON operating_expenses(plan_id)`},

	// Migration 37: backfill operating_expenses (and its wizard_sections
	// completion row) from every existing plan's plans.data JSON blob,
	// since migration 35 introduced a new table that must not orphan
	// pre-existing Operating Expenses data.
	{Fn: backfillOperatingExpensesData},

	// Migrations 38-47: add deleted_at column (soft-delete support) to all
	// business entity tables. Per the Repository Rules, hard DELETE is
	// forbidden; instead each row is marked deleted_at = unix-epoch-seconds.
	// Singleton tables (starting_balances, payroll_tax_rates, sales_growth_curve)
	// and join/marker tables (plan_access, wizard_sections) keep hard-delete
	// because they have no individual-row audit value.
	{SQL: `ALTER TABLE plans ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE capital_assets ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE startup_costs ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE funding_sources ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE salary_roles ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE benefits ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE products ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE inventory_purchases ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE distributions ADD COLUMN deleted_at INTEGER`},
	{SQL: `ALTER TABLE operating_expenses ADD COLUMN deleted_at INTEGER`},

	// Migration 48: Transactional outbox for domain events.
	// Repositories write domain events here in the same transaction that
	// saves the aggregate; a relay goroutine reads and delivers them
	// separately, guaranteeing at-least-once delivery even if the process
	// crashes after the commit.
	{SQL: `CREATE TABLE IF NOT EXISTS outbox_events (
		id         TEXT PRIMARY KEY,
		event_name TEXT NOT NULL,
		payload    TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		processed_at INTEGER
	)`},
	{SQL: `CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON outbox_events(created_at) WHERE processed_at IS NULL`},

	// Migration 49: Idempotency key store.
	// Every mutating command reserves a key here with status='running'
	// before executing; on completion it caches the JSON response and
	// flips to 'completed'. A second request with the same key returns
	// the cached response immediately without re-executing.
	{SQL: `CREATE TABLE IF NOT EXISTS idempotency_keys (
		key        TEXT PRIMARY KEY,
		status     TEXT NOT NULL DEFAULT 'running',
		response   TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`},

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

// markWizardSectionComplete is markSectionComplete's wizard_sections
// equivalent, used by migrations that run after wizard_sections replaces
// starting_point_sections (migration 33 onward).
func markWizardSectionComplete(tx *sql.Tx, planID, hub, section string, completedAt int64) error {
	_, err := tx.Exec(
		`INSERT INTO wizard_sections (plan_id, hub, section, completed_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(plan_id, hub, section) DO UPDATE SET completed_at = excluded.completed_at`,
		planID, hub, section, completedAt,
	)
	if err != nil {
		return fmt.Errorf("backfill: failed to mark section %s/%s complete for plan %s: %w", hub, section, planID, err)
	}
	return nil
}

// migrateStartingPointSectionsToWizardSections copies every row from the
// Starting-Point-specific starting_point_sections table into the new
// generic wizard_sections table (hub='starting-point'), then drops the old
// table, since Payroll/Sales Forecast/Cash Flow need the identical
// completion-tracking mechanism and one table per hub doesn't scale.
func migrateStartingPointSectionsToWizardSections(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("wizard_sections migration: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT plan_id, section, completed_at FROM starting_point_sections")
	if err != nil {
		return fmt.Errorf("wizard_sections migration: failed to query starting_point_sections: %w", err)
	}

	type legacySectionRow struct {
		planID      string
		section     string
		completedAt sql.NullInt64
	}
	var legacyRows []legacySectionRow
	for rows.Next() {
		var r legacySectionRow
		if err := rows.Scan(&r.planID, &r.section, &r.completedAt); err != nil {
			rows.Close()
			return fmt.Errorf("wizard_sections migration: failed to scan row: %w", err)
		}
		legacyRows = append(legacyRows, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("wizard_sections migration: failed to iterate rows: %w", err)
	}
	rows.Close()

	for _, r := range legacyRows {
		if _, err := tx.Exec(
			`INSERT INTO wizard_sections (plan_id, hub, section, completed_at) VALUES (?, ?, ?, ?)
			 ON CONFLICT(plan_id, hub, section) DO UPDATE SET completed_at = excluded.completed_at`,
			r.planID, domain.HubStartingPoint, r.section, r.completedAt,
		); err != nil {
			return fmt.Errorf("wizard_sections migration: failed to insert row for plan %s: %w", r.planID, err)
		}
	}

	if _, err := tx.Exec("DROP TABLE starting_point_sections"); err != nil {
		return fmt.Errorf("wizard_sections migration: failed to drop starting_point_sections: %w", err)
	}

	return tx.Commit()
}

// legacyPlanPayrollSalesCashFlow mirrors the subset of domain.Plan's old
// planJSON fields relevant to Payroll/Sales Forecast/Cash Flow, for
// decoding the pre-normalization JSON blob during backfill. Field names
// match the old planJSON's json tags exactly; nested struct field names
// have no json tags in the domain package, so they marshal/unmarshal using
// their literal Go names.
type legacyPlanPayrollSalesCashFlow struct {
	SalaryRoles []struct {
		Role           string
		IsContractor   bool
		Headcount      int
		MonthlyPay     int64
		GrowthAfterYr1 legacyAnnualGrowth
	} `json:"salaryRoles"`
	Benefits []struct {
		Type           string
		MonthlyAmount  int64
		GrowthAfterYr1 legacyAnnualGrowth
	} `json:"benefits"`
	PayrollTaxRates struct {
		SocialSecurityRate float64
		MedicareRate       float64
		FUTARate           float64
		SUTARate           float64
	} `json:"payrollTaxRates"`
	Products []struct {
		Name         string
		Month1Units  int
		PricePerUnit int64
		CostPerUnit  int64
	} `json:"products"`
	SalesGrowth struct {
		Year1QuarterlyRates [4]float64
		FutureYearRates     []float64
	} `json:"salesGrowth"`
	AdditionalInventory []struct {
		Category       string
		MonthlyAmount  int64
		GrowthAfterYr1 legacyAnnualGrowth
	} `json:"additionalInventory"`
	Distributions []struct {
		Name           string
		MonthlyAmount  int64
		GrowthAfterYr1 legacyAnnualGrowth
	} `json:"distributions"`
}

type legacyAnnualGrowth struct {
	RatesAfterYear1 []float64
}

// yr2Yr3 extracts the Year 2 / Year 3 rates from a legacy AnnualGrowth's
// (potentially longer, or shorter) RatesAfterYear1 slice, matching the new
// fixed growth_yr2/growth_yr3 columns every entity table now uses.
func (g legacyAnnualGrowth) yr2Yr3() (float64, float64) {
	var yr2, yr3 float64
	if len(g.RatesAfterYear1) > 0 {
		yr2 = g.RatesAfterYear1[0]
	}
	if len(g.RatesAfterYear1) > 1 {
		yr3 = g.RatesAfterYear1[1]
	}
	return yr2, yr3
}

// backfillPayrollSalesForecastCashFlowData populates the new normalized
// tables from every existing plan's JSON blob so pre-existing Payroll/Sales
// Forecast/Cash Flow data isn't orphaned by migrations 20-32. Mirrors
// backfillStartingPointData's structure and tradeoffs.
func backfillPayrollSalesForecastCashFlowData(db *sql.DB) error {
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
		var legacy legacyPlanPayrollSalesCashFlow
		if err := json.Unmarshal(pr.data, &legacy); err != nil {
			return fmt.Errorf("backfill: failed to decode plan %s: %w", pr.id, err)
		}

		now := pr.updatedAt

		if len(legacy.SalaryRoles) > 0 {
			for i, s := range legacy.SalaryRoles {
				yr2, yr3 := s.GrowthAfterYr1.yr2Yr3()
				if _, err := tx.Exec(
					`INSERT INTO salary_roles (id, plan_id, role, is_contractor, headcount, monthly_pay, growth_yr2, growth_yr3, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, s.Role, s.IsContractor, s.Headcount, s.MonthlyPay, yr2, yr3, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert salary_role for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubPayroll, domain.SectionSalaryRoles, now); err != nil {
				return err
			}
		}

		if len(legacy.Benefits) > 0 {
			for i, b := range legacy.Benefits {
				yr2, yr3 := b.GrowthAfterYr1.yr2Yr3()
				if _, err := tx.Exec(
					`INSERT INTO benefits (id, plan_id, type, monthly_amount, growth_yr2, growth_yr3, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, b.Type, b.MonthlyAmount, yr2, yr3, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert benefit for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubPayroll, domain.SectionBenefits, now); err != nil {
				return err
			}
		}

		tr := legacy.PayrollTaxRates
		if tr.SocialSecurityRate != 0 || tr.MedicareRate != 0 || tr.FUTARate != 0 || tr.SUTARate != 0 {
			if _, err := tx.Exec(
				`INSERT INTO payroll_tax_rates (plan_id, social_security_rate, medicare_rate, futa_rate, suta_rate, current_step, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, 0, ?, ?)`,
				pr.id, tr.SocialSecurityRate, tr.MedicareRate, tr.FUTARate, tr.SUTARate, now, now,
			); err != nil {
				return fmt.Errorf("backfill: failed to insert payroll_tax_rates for plan %s: %w", pr.id, err)
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubPayroll, domain.SectionPayrollTaxRates, now); err != nil {
				return err
			}
		}

		if len(legacy.Products) > 0 {
			for i, p := range legacy.Products {
				if _, err := tx.Exec(
					`INSERT INTO products (id, plan_id, name, month1_units, price_per_unit, cost_per_unit, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, p.Name, p.Month1Units, p.PricePerUnit, p.CostPerUnit, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert product for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubSalesForecast, domain.SectionProducts, now); err != nil {
				return err
			}
		}

		sg := legacy.SalesGrowth
		sgHasData := sg.Year1QuarterlyRates[0] != 0 || sg.Year1QuarterlyRates[1] != 0 || sg.Year1QuarterlyRates[2] != 0 || sg.Year1QuarterlyRates[3] != 0 || len(sg.FutureYearRates) > 0
		if sgHasData {
			var growthYr2, growthYr3 float64
			if len(sg.FutureYearRates) > 0 {
				growthYr2 = sg.FutureYearRates[0]
			}
			if len(sg.FutureYearRates) > 1 {
				growthYr3 = sg.FutureYearRates[1]
			}
			if _, err := tx.Exec(
				`INSERT INTO sales_growth_curve (plan_id, y1_q1, y1_q2, y1_q3, y1_q4, growth_yr2, growth_yr3, current_step, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
				pr.id, sg.Year1QuarterlyRates[0], sg.Year1QuarterlyRates[1], sg.Year1QuarterlyRates[2], sg.Year1QuarterlyRates[3], growthYr2, growthYr3, now, now,
			); err != nil {
				return fmt.Errorf("backfill: failed to insert sales_growth_curve for plan %s: %w", pr.id, err)
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubSalesForecast, domain.SectionSalesGrowthCurve, now); err != nil {
				return err
			}
		}

		if len(legacy.AdditionalInventory) > 0 {
			for i, inv := range legacy.AdditionalInventory {
				yr2, yr3 := inv.GrowthAfterYr1.yr2Yr3()
				if _, err := tx.Exec(
					`INSERT INTO inventory_purchases (id, plan_id, category, monthly_amount, growth_yr2, growth_yr3, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, inv.Category, inv.MonthlyAmount, yr2, yr3, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert inventory_purchase for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubCashFlow, domain.SectionInventoryPurchases, now); err != nil {
				return err
			}
		}

		if len(legacy.Distributions) > 0 {
			for i, d := range legacy.Distributions {
				yr2, yr3 := d.GrowthAfterYr1.yr2Yr3()
				if _, err := tx.Exec(
					`INSERT INTO distributions (id, plan_id, name, monthly_amount, growth_yr2, growth_yr3, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, d.Name, d.MonthlyAmount, yr2, yr3, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert distribution for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubCashFlow, domain.SectionDistributions, now); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// legacyPlanOperatingExpenses mirrors the subset of domain.Plan's old
// planJSON fields relevant to Operating Expenses, for decoding the
// pre-normalization JSON blob during backfill.
type legacyPlanOperatingExpenses struct {
	OpEx []struct {
		Name               string
		BaseAmountPerMonth int64
		Growth             struct {
			Type       string
			AnnualRate float64
		}
	} `json:"opEx"`
}

// backfillOperatingExpensesData populates the operating_expenses table
// from every existing plan's JSON blob so pre-existing Operating Expenses
// data isn't orphaned by migration 35. Mirrors
// backfillStartingPointData's structure and tradeoffs.
func backfillOperatingExpensesData(db *sql.DB) error {
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
		var legacy legacyPlanOperatingExpenses
		if err := json.Unmarshal(pr.data, &legacy); err != nil {
			return fmt.Errorf("backfill: failed to decode plan %s: %w", pr.id, err)
		}

		now := pr.updatedAt

		if len(legacy.OpEx) > 0 {
			for i, e := range legacy.OpEx {
				growthType := e.Growth.Type
				if growthType == "" {
					growthType = "Flat"
				}
				if _, err := tx.Exec(
					`INSERT INTO operating_expenses (id, plan_id, name, monthly_amount, growth_type, annual_rate, status, current_step, sort_order, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, 'complete', 0, ?, ?, ?)`,
					uuid.New().String(), pr.id, e.Name, e.BaseAmountPerMonth, growthType, e.Growth.AnnualRate, i, now, now,
				); err != nil {
					return fmt.Errorf("backfill: failed to insert operating_expense for plan %s: %w", pr.id, err)
				}
			}
			if err := markWizardSectionComplete(tx, pr.id, domain.HubOperatingExpenses, domain.SectionOperatingExpenses, now); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
