package main

import (
	"fmt"
	"log"

	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

func migrate(dbPath string) {
	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer conn.Close()

	if err := sqlite.RunMigrations(conn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	fmt.Println("✓ Database migrations completed successfully")
}
