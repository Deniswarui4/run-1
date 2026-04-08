//go:build integration

package migrations_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/warui/event-ticketing-api/db/migrations"
)

// TestInvalidMigrationFails verifies that goose returns a non-nil error when
// a migration file contains invalid SQL.
//
// Validates: Requirements 7.8
func TestInvalidMigrationFails(t *testing.T) {
	db := testDB(t)

	// Restore the real embedded FS when done so other tests are unaffected.
	t.Cleanup(func() {
		goose.SetBaseFS(migrations.FS)
	})

	// Create a temp dir with a single bad migration file.
	tmpDir := t.TempDir()
	badSQL := `-- +goose Up
THIS IS NOT VALID SQL;
-- +goose Down
-- nothing
`
	badFile := filepath.Join(tmpDir, "00001_bad.sql")
	if err := os.WriteFile(badFile, []byte(badSQL), 0o644); err != nil {
		t.Fatalf("failed to write bad migration file: %v", err)
	}

	// Switch goose to use the OS filesystem pointing at our temp dir.
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	err := goose.Up(db, tmpDir)
	if err == nil {
		t.Fatal("expected goose.Up to fail on invalid SQL, but it succeeded")
	}

	if err.Error() == "" {
		t.Error("expected a non-empty error message from goose.Up with invalid SQL")
	}
}

// TestUnreachableDBFails verifies that attempting to connect to an unreachable
// database host returns a non-nil, descriptive error.
//
// Validates: Requirements 7.9
func TestUnreachableDBFails(t *testing.T) {
	dsn := fmt.Sprintf(
		"host=127.0.0.1 port=19999 user=postgres password=postgres dbname=nonexistent sslmode=disable connect_timeout=2",
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// sql.Open rarely errors; if it does the error is already descriptive.
		return
	}
	defer db.Close()

	pingErr := db.Ping()
	if pingErr == nil {
		t.Fatal("expected db.Ping to fail for unreachable host, but it succeeded")
	}

	if pingErr.Error() == "" {
		t.Error("expected a non-empty error message when connecting to unreachable DB")
	}
}
