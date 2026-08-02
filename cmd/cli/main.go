package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func serveCmd() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbPath := fs.String("db", "./northbasis.db", "path to SQLite database file")
	port := fs.String("port", ":8080", "port to listen on")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	serve(*dbPath, *port)
}

func migrateCmd() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	dbPath := fs.String("db", "./northbasis.db", "path to SQLite database file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	migrate(*dbPath)
}

func resetCmd() {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	dbPath := fs.String("db", "./northbasis.db", "path to SQLite database file")
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
  help        Show this help message

Examples:
  %s serve --db ./northbasis.db --port :8080
  %s migrate --db ./northbasis.db
  %s reset --db ./northbasis.db

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
