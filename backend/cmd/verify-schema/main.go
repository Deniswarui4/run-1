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

	fmt.Println("Checking schema for migration 00003_guest_cart_featured.sql...")
	fmt.Println()

	// Check all expected tables and columns from migration 00003
	checks := []struct {
		name  string
		query string
	}{
		{
			name: "guest_sessions table",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'guest_sessions'
			)`,
		},
		{
			name: "tickets.guest_email column",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'tickets' AND column_name = 'guest_email'
			)`,
		},
		{
			name: "draft_carts table",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'draft_carts'
			)`,
		},
		{
			name: "draft_cart_items table",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'draft_cart_items'
			)`,
		},
		{
			name: "events.featured_type column",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'events' AND column_name = 'featured_type'
			)`,
		},
		{
			name: "platform_settings.auto_feature_threshold column",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'platform_settings' AND column_name = 'auto_feature_threshold'
			)`,
		},
		{
			name: "event_feature_evaluations table",
			query: `SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'event_feature_evaluations'
			)`,
		},
	}

	allGood := true
	for _, check := range checks {
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
			allGood = false
		}
	}

	fmt.Println()
	if allGood {
		fmt.Println("✓ All schema elements from migration 00003 are present!")
	} else {
		fmt.Println("✗ Some schema elements are missing. The migration may not have been fully applied.")
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
