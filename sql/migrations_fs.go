// Package sql embeds the sql-migrate migration files so the compiled binary
// carries them along, matching internal/views' embed.FS pattern for
// templates/static assets rather than depending on relative file paths at
// runtime.
package sql

import "embed"

//go:embed migrations/*.sql
var MigrationsFS embed.FS
