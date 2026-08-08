package sqlite

import (
	"database/sql"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
	sqlembed "github.com/zaidmasri/business-planning-tool/sql"
)

func init() {
	// gorp_migrations is sql-migrate's default table name; set explicitly so
	// it never collides with the retired hand-rolled schema_migrations table.
	migrate.SetTable("gorp_migrations")
}

// RunMigrations applies every pending migration embedded in
// sql/migrations_fs.go. Safe to call on every process start (no-op once the
// schema is current).
func RunMigrations(db *sql.DB) error {
	source := migrate.EmbedFileSystemMigrationSource{
		FileSystem: sqlembed.MigrationsFS,
		Root:       "migrations",
	}

	if _, err := migrate.Exec(db, "sqlite3", source, migrate.Up); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
