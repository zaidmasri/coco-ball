// Package sqlite implements the domain repository interfaces
// (internal/domain/repositories) on top of the sqlc-generated query code in
// internal/infrastructure/db/sqlc, backed by SQLite via mattn/go-sqlite3.
package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// NewConnection opens a SQLite database file and verifies it's reachable.
func NewConnection(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
