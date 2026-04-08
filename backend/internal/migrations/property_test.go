//go:build integration

package migrations_test

import (
	"database/sql"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

// schemaSnapshot returns a map of public BASE TABLE names present in the DB.
func schemaSnapshot(t *testing.T, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		t.Fatalf("schemaSnapshot: query failed: %v", err)
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("schemaSnapshot: scan failed: %v", err)
		}
		tables[name] = true
	}
	return tables
}

// snapshotsEqual returns true when both maps contain exactly the same keys.
func snapshotsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// TestMigrationRoundTrip verifies Property 1: for any migration version,
// applying Up then Down leaves the schema identical to the pre-Up state.
//
// Validates: Requirements 7.1
func TestMigrationRoundTrip(t *testing.T) {
	db := testDB(t)

	// Collect available migration versions (0 → MaxInt64 = all migrations).
	ms, err := goose.CollectMigrations(".", 0, math.MaxInt64)
	if err != nil {
		t.Fatalf("failed to collect migrations: %v", err)
	}
	if len(ms) == 0 {
		t.Fatal("no migrations found")
	}

	versions := make([]int64, len(ms))
	for i, m := range ms {
		versions[i] = m.Version
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const iterations = 100
	for i := 0; i < iterations; i++ {
		version := versions[rng.Intn(len(versions))]

		// 1. Bring DB to the state just before this migration.
		if version == versions[0] {
			// First migration: pre-Up state is an empty DB.
			if err := goose.DownTo(db, ".", 0); err != nil {
				t.Fatalf("iter %d: DownTo(0) failed: %v", i, err)
			}
		} else {
			if err := goose.UpTo(db, ".", version-1); err != nil {
				t.Fatalf("iter %d: UpTo(%d) failed: %v", i, version-1, err)
			}
		}

		// 2. Snapshot schema before applying the migration.
		preSnapshot := schemaSnapshot(t, db)

		// 3. Apply the migration.
		if err := goose.UpTo(db, ".", version); err != nil {
			t.Fatalf("iter %d: UpTo(%d) failed: %v", i, version, err)
		}

		// 4. Roll back the last applied migration.
		if err := goose.Down(db, "."); err != nil {
			t.Fatalf("iter %d: Down failed: %v", i, err)
		}

		// 5. Snapshot schema after rollback.
		postSnapshot := schemaSnapshot(t, db)

		// 6. Assert the schema is unchanged.
		if !snapshotsEqual(preSnapshot, postSnapshot) {
			t.Errorf("iter %d (version %d): schema mismatch after round-trip\n  pre:  %v\n  post: %v",
				i, version, preSnapshot, postSnapshot)
		}

		// 7. Reset to clean state for the next iteration.
		if err := goose.DownTo(db, ".", 0); err != nil {
			t.Fatalf("iter %d: cleanup DownTo(0) failed: %v", i, err)
		}
	}
}
