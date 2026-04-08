//go:build integration

package migrations_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/warui/event-ticketing-api/db/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "event_ticketing_test")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("failed to ping DB: %v", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		db.Close()
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	// Reset to clean state
	if err := goose.DownTo(db, ".", 0); err != nil {
		db.Close()
		t.Fatalf("failed to reset DB: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

var appTables = []string{
	"users",
	"categories",
	"events",
	"ticket_types",
	"tickets",
	"transactions",
	"platform_settings",
	"withdrawal_requests",
	"organizer_balances",
}

func getPublicTables(t *testing.T, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		t.Fatalf("failed to query information_schema.tables: %v", err)
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		tables[name] = true
	}
	return tables
}

// TestMigrationsUp applies all migrations from an empty DB and asserts all 9
// application tables exist.
//
// Validates: Requirements 7.2
func TestMigrationsUp(t *testing.T) {
	db := testDB(t)

	if err := goose.Up(db, "."); err != nil {
		t.Fatalf("goose.Up failed: %v", err)
	}

	tables := getPublicTables(t, db)

	for _, tbl := range appTables {
		if !tables[tbl] {
			t.Errorf("expected table %q to exist after Up, but it was not found", tbl)
		}
	}
}

// TestMigrationsUpThenDown applies all migrations then rolls them all back and
// asserts none of the 9 application tables remain (Property 2).
//
// Validates: Requirements 7.3
func TestMigrationsUpThenDown(t *testing.T) {
	db := testDB(t)

	if err := goose.Up(db, "."); err != nil {
		t.Fatalf("goose.Up failed: %v", err)
	}

	if err := goose.DownTo(db, ".", 0); err != nil {
		t.Fatalf("goose.DownTo(0) failed: %v", err)
	}

	tables := getPublicTables(t, db)

	for _, tbl := range appTables {
		if tables[tbl] {
			t.Errorf("expected table %q to be gone after DownTo(0), but it still exists", tbl)
		}
	}
}
