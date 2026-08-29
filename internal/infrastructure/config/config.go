// Package config centralizes the application's environment-driven settings.
package config

import "os"

// Config holds settings the app needs to start: where the SQLite database
// file lives, which port the HTTP server listens on, and how the worker
// sends email.
type Config struct {
	DBPath string
	Port   string

	// AppBaseURL is linked to from transactional emails (welcome, invite).
	AppBaseURL string

	// SMTP settings for the worker's mailer. SMTPHost empty means "no SMTP
	// configured" — the worker falls back to logging emails instead of
	// sending them, so local dev needs no mail server.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Load reads settings from the environment, falling back to the same
// defaults cmd/cli's flags have always used (./northbasis.db, :8080).
func Load() Config {
	return Config{
		DBPath:       getEnv("DB_PATH", "./northbasis.db"),
		Port:         getEnv("PORT", ":8080"),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:8080"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "Northbasis <no-reply@northbasis.com>"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
