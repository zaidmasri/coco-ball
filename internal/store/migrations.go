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
}

// GetMigrations returns all migration SQL statements
func GetMigrations() []string {
	return migrations
}
