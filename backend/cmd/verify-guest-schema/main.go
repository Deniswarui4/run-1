package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("Checking guest checkout schema requirements...")
	fmt.Println()

	checks := []struct {
		name  string
		query string
		fix   string
	}{
		{
			name: "tickets.guest_email column",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'tickets' AND column_name = 'guest_email'
			)`,
			fix: "ALTER TABLE tickets ADD COLUMN guest_email VARCHAR;",
		},
		{
			name: "tickets.attendee_id is nullable",
			query: `SELECT is_nullable FROM information_schema.columns 
				WHERE table_name = 'tickets' AND column_name = 'attendee_id'`,
			fix: "ALTER TABLE tickets ALTER COLUMN attendee_id DROP NOT NULL;",
		},
		{
			name: "guest_sessions table",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'guest_sessions'
			)`,
			fix: "Run migration 00003_guest_cart_featured.sql",
		},
	}

	allGood := true

	for _, check := range checks {
		if check.name == "tickets.attendee_id is nullable" {
			var isNullable string
			err := db.QueryRow(check.query).Scan(&isNullable)
			if err != nil {
				fmt.Printf("✗ Error checking %s: %v\n", check.name, err)
				allGood = false
				continue
			}

			if isNullable == "YES" {
				fmt.Printf("✓ %s\n", check.name)
			} else {
				fmt.Printf("✗ %s - attendee_id is NOT NULL (should be nullable)\n", check.name)
				fmt.Printf("  Fix: %s\n", check.fix)
				allGood = false
			}
		} else {
			var exists bool
			err := db.QueryRow(check.query).Scan(&exists)
			if err != nil {
				fmt.Printf("✗ Error checking %s: %v\n", check.name, err)
				allGood = false
				continue
			}

			if exists {
				fmt.Printf("✓ %s exists\n", check.name)
			} else {
				fmt.Printf("✗ %s MISSING\n", check.name)
				fmt.Printf("  Fix: %s\n", check.fix)
				allGood = false
			}
		}
	}

	fmt.Println()

	// Check for any existing guest tickets
	var guestTicketCount int
	db.QueryRow("SELECT COUNT(*) FROM tickets WHERE guest_email IS NOT NULL").Scan(&guestTicketCount)
	fmt.Printf("Existing guest tickets in database: %d\n", guestTicketCount)

	// Check for transactions with nil user_id (guest transactions)
	var guestTransactionCount int
	db.QueryRow("SELECT COUNT(*) FROM transactions WHERE user_id = '00000000-0000-0000-0000-000000000000'").Scan(&guestTransactionCount)
	fmt.Printf("Guest transactions (user_id = nil UUID): %d\n", guestTransactionCount)

	fmt.Println()

	if allGood {
		fmt.Println("✓ All guest checkout schema requirements are met!")
		fmt.Println("\nIf you're still experiencing errors, please:")
		fmt.Println("1. Start the backend server: go run ./cmd/api")
		fmt.Println("2. Check the server logs for specific error messages")
		fmt.Println("3. Verify Paystack credentials are configured correctly")
		fmt.Println("4. Ensure events have active ticket types")
	} else {
		fmt.Println("✗ Some schema requirements are missing.")
		fmt.Println("\nRun: go run ./cmd/migrate up")
		os.Exit(1)
	}
}

func openDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "event_ticketing")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to reach database at %s:%s — %w", host, port, err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
