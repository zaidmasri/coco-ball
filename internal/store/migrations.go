package store

var migrations = []string{
	// Migration 1: Create migrations tracking table (MUST BE FIRST)
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`,

	// Migration 2: Create users table
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		created_at INTEGER NOT NULL
	)`,

	// Migration 3: Create users_credentials table
	`CREATE TABLE IF NOT EXISTS users_credentials (
		email TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE
	)`,

	// Migration 4: Create sessions table
	`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`,

	// Migration 5: Create plans table
	`CREATE TABLE IF NOT EXISTS plans (
		id TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`,

	// Migration 6: Create plan_access table
	`CREATE TABLE IF NOT EXISTS plan_access (
		plan_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		access_level TEXT NOT NULL,
		invited_at INTEGER NOT NULL,
		PRIMARY KEY (plan_id, user_id),
		FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`,

	// Migration 7: Add first name to users
	`ALTER TABLE users ADD COLUMN first_name TEXT NOT NULL DEFAULT ''`,

	// Migration 8: Add last name to users
	`ALTER TABLE users ADD COLUMN last_name TEXT NOT NULL DEFAULT ''`,

	// Migration 9: Create plan_invites table
	`CREATE TABLE IF NOT EXISTS plan_invites (
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
	)`,

	// Migration 10: Index invites by email for onboarding lookups
	`CREATE INDEX IF NOT EXISTS idx_plan_invites_email ON plan_invites(email)`,
}

// GetMigrations returns all migration SQL statements
func GetMigrations() []string {
	return migrations
}
