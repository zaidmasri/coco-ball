package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		serveCmd()
	case "migrate":
		migrateCmd()
	case "reset":
		resetCmd()
	case "worker":
		workerCmd()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func serveCmd() {
	cfg := config.Load()
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "path to SQLite database file")
	port := fs.String("port", cfg.Port, "port to listen on")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	serve(*dbPath, *port)
}

func migrateCmd() {
	cfg := config.Load()
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "path to SQLite database file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	migrate(*dbPath)
}

func workerCmd() {
	cfg := config.Load()
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "path to SQLite database file")
	interval := fs.Duration("interval", 5*time.Second, "outbox poll interval")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	worker(*dbPath, *interval, cfg)
}

func resetCmd() {
	cfg := config.Load()
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "path to SQLite database file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	if err := os.Remove(*dbPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove database: %v", err)
	}
	fmt.Printf("Database reset. Run 'migrate' to recreate the schema.\n")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `
Usage: %s <command> [options]

Commands:
  serve       Start the web application
  migrate     Run database migrations
  reset       Reset and delete the database
  worker      Run the background worker that consumes outbox events (emails)
  help        Show this help message

Examples:
  %s serve --db ./northbasis.db --port :8080
  %s migrate --db ./northbasis.db
  %s reset --db ./northbasis.db
  %s worker --db ./northbasis.db --interval 5s

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
