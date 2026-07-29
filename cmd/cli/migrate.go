package main

import (
	"fmt"
	"log"

	"github.com/zaidmasri/business-planning-tool/internal/store"
)

func migrate(dbPath string) {
	// Initialize store which automatically runs migrations
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer sqliteStore.Close()

	fmt.Println("✓ Database migrations completed successfully")
}
