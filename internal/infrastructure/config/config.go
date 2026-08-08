// Package config centralizes the application's environment-driven settings.
package config

import "os"

// Config holds settings the app needs to start: where the SQLite database
// file lives and which port the HTTP server listens on.
type Config struct {
	DBPath string
	Port   string
}

// Load reads settings from the environment, falling back to the same
// defaults cmd/cli's flags have always used (./northbasis.db, :8080).
func Load() Config {
	return Config{
		DBPath: getEnv("DB_PATH", "./northbasis.db"),
		Port:   getEnv("PORT", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
