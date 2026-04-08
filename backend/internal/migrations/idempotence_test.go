//go:build integration

package migrations_test

import (
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/warui/event-ticketing-api/db/migrations"
)

// TestMigrationIdempotence verifies Property 3: applying goose.Up on an
// already-fully-migrated DB does not change the schema or the
// goose_db_version row count.
//
// Validates: Requirements 7.5
func TestMigrationIdempotence(t *testing.T) {
	db := testDB(t)

	const iterations = 100
	for i := 0; i < iterations; i++ {
		// Reset to clean state at the start of each iteration.
		if err := goose.DownTo(db, ".", 0); err != nil {
			t.Fatalf("iter %d: DownTo(0) failed: %v", i, err)
		}

		// Apply all migrations.
		if err := goose.Up(db, "."); err != nil {
			t.Fatalf("iter %d: first goose.Up failed: %v", i, err)
		}

		// Count applied rows in goose_db_version.
		var countBefore int
		if err := db.QueryRow(`SELECT COUNT(*) FROM goose_db_version WHERE is_applied = true`).Scan(&countBefore); err != nil {
			t.Fatalf("iter %d: count before failed: %v", i, err)
		}

		// Snapshot schema before second Up.
		schemaBefore := schemaSnapshot(t, db)

		// Run Up again — should be a no-op.
		if err := goose.Up(db, "."); err != nil {
			t.Fatalf("iter %d: second goose.Up failed: %v", i, err)
		}

		// Count again.
		var countAfter int
		if err := db.QueryRow(`SELECT COUNT(*) FROM goose_db_version WHERE is_applied = true`).Scan(&countAfter); err != nil {
			t.Fatalf("iter %d: count after failed: %v", i, err)
		}

		// Snapshot schema after second Up.
		schemaAfter := schemaSnapshot(t, db)

		if countBefore != countAfter {
			t.Errorf("iter %d: goose_db_version count changed: before=%d after=%d", i, countBefore, countAfter)
		}
		if !snapshotsEqual(schemaBefore, schemaAfter) {
			t.Errorf("iter %d: schema changed after second Up\n  before: %v\n  after:  %v", i, schemaBefore, schemaAfter)
		}
	}
}

// extractUpSQL reads the Up section from 00002_seed_data.sql, stripping goose
// directive lines, and returns the raw SQL ready for direct execution.
func extractUpSQL(t *testing.T) string {
	t.Helper()

	data, err := migrations.FS.ReadFile("00002_seed_data.sql")
	if err != nil {
		t.Fatalf("extractUpSQL: failed to read seed file: %v", err)
	}

	content := string(data)

	// Take everything before "-- +goose Down".
	if idx := strings.Index(content, "-- +goose Down"); idx >= 0 {
		content = content[:idx]
	}

	// Strip goose directive lines.
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose") {
			continue
		}
		lines = append(lines, line)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// TestSeedDataIdempotence verifies Properties 6.2 + 6.3: running the seed SQL
// multiple times always results in exactly 1 platform_settings row and exactly
// 10 categories rows.
//
// Validates: Requirements 7.6
func TestSeedDataIdempotence(t *testing.T) {
	db := testDB(t)

	// Apply schema migration only (version 1).
	if err := goose.UpTo(db, ".", 1); err != nil {
		t.Fatalf("goose.UpTo(1) failed: %v", err)
	}

	seedSQL := extractUpSQL(t)

	const iterations = 100
	for i := 0; i < iterations; i++ {
		if _, err := db.Exec(seedSQL); err != nil {
			t.Fatalf("iter %d: seed SQL exec failed: %v", i, err)
		}

		var platformCount int
		if err := db.QueryRow(`SELECT COUNT(*) FROM platform_settings`).Scan(&platformCount); err != nil {
			t.Fatalf("iter %d: count platform_settings failed: %v", i, err)
		}
		if platformCount != 1 {
			t.Errorf("iter %d: expected 1 platform_settings row, got %d", i, platformCount)
		}

		var categoryCount int
		if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&categoryCount); err != nil {
			t.Fatalf("iter %d: count categories failed: %v", i, err)
		}
		if categoryCount != 10 {
			t.Errorf("iter %d: expected 10 categories rows, got %d", i, categoryCount)
		}
	}
}
