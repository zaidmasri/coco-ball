-- +migrate Up

CREATE TABLE users (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	created_at INTEGER NOT NULL,
	first_name TEXT NOT NULL DEFAULT '',
	last_name TEXT NOT NULL DEFAULT ''
);

CREATE TABLE users_credentials (
	email TEXT PRIMARY KEY,
	password_hash TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE
);

CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE plans (
	id TEXT PRIMARY KEY,
	data BLOB NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);

CREATE TABLE plan_access (
	plan_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	access_level TEXT NOT NULL,
	invited_at INTEGER NOT NULL,
	PRIMARY KEY (plan_id, user_id),
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE plan_invites (
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
);

CREATE INDEX idx_plan_invites_email ON plan_invites(email);

CREATE TABLE capital_assets (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_capital_assets_plan_id ON capital_assets(plan_id);

CREATE TABLE startup_costs (
	id TEXT PRIMARY KEY,
	plan_id TEXT NOT NULL,
	name TEXT NOT NULL DEFAULT '',
	amount INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'draft',
	current_step INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_startup_costs_plan_id ON startup_costs(plan_id);

CREATE TABLE funding_sources (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_funding_sources_plan_id ON funding_sources(plan_id);

CREATE TABLE starting_balances (
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
);

CREATE TABLE salary_roles (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_salary_roles_plan_id ON salary_roles(plan_id);

CREATE TABLE benefits (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_benefits_plan_id ON benefits(plan_id);

CREATE TABLE payroll_tax_rates (
	plan_id TEXT PRIMARY KEY,
	social_security_rate REAL NOT NULL DEFAULT 0,
	medicare_rate REAL NOT NULL DEFAULT 0,
	futa_rate REAL NOT NULL DEFAULT 0,
	suta_rate REAL NOT NULL DEFAULT 0,
	current_step INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE TABLE products (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_products_plan_id ON products(plan_id);

CREATE TABLE sales_growth_curve (
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
);

CREATE TABLE inventory_purchases (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_inventory_purchases_plan_id ON inventory_purchases(plan_id);

CREATE TABLE distributions (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_distributions_plan_id ON distributions(plan_id);

CREATE TABLE wizard_sections (
	plan_id TEXT NOT NULL,
	hub TEXT NOT NULL,
	section TEXT NOT NULL,
	completed_at INTEGER,
	PRIMARY KEY (plan_id, hub, section),
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE TABLE operating_expenses (
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
	deleted_at INTEGER,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_operating_expenses_plan_id ON operating_expenses(plan_id);

CREATE TABLE outbox_events (
	id TEXT PRIMARY KEY,
	event_name TEXT NOT NULL,
	payload TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	processed_at INTEGER
);

CREATE INDEX idx_outbox_unprocessed ON outbox_events(created_at) WHERE processed_at IS NULL;

CREATE TABLE idempotency_keys (
	key TEXT PRIMARY KEY,
	status TEXT NOT NULL DEFAULT 'running',
	response TEXT,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

-- +migrate Down

DROP TABLE idempotency_keys;
DROP TABLE outbox_events;
DROP INDEX idx_operating_expenses_plan_id;
DROP TABLE operating_expenses;
DROP TABLE wizard_sections;
DROP INDEX idx_distributions_plan_id;
DROP TABLE distributions;
DROP INDEX idx_inventory_purchases_plan_id;
DROP TABLE inventory_purchases;
DROP TABLE sales_growth_curve;
DROP INDEX idx_products_plan_id;
DROP TABLE products;
DROP TABLE payroll_tax_rates;
DROP INDEX idx_benefits_plan_id;
DROP TABLE benefits;
DROP INDEX idx_salary_roles_plan_id;
DROP TABLE salary_roles;
DROP TABLE starting_balances;
DROP INDEX idx_funding_sources_plan_id;
DROP TABLE funding_sources;
DROP INDEX idx_startup_costs_plan_id;
DROP TABLE startup_costs;
DROP INDEX idx_capital_assets_plan_id;
DROP TABLE capital_assets;
DROP INDEX idx_plan_invites_email;
DROP TABLE plan_invites;
DROP TABLE plan_access;
DROP TABLE plans;
DROP TABLE sessions;
DROP TABLE users_credentials;
DROP TABLE users;
