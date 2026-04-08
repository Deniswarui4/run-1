// Package migrations embeds the goose SQL migration files so they can be
// bundled into any binary that needs to run migrations.
package migrations

import "embed"

// FS holds all versioned SQL migration files.
//
//go:embed *.sql
var FS embed.FS
