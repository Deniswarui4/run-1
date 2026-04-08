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

	// Check if featured_type column exists
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'events' 
			AND column_name = 'featured_type'
		)
	`
	
	err = db.QueryRow(checkQuery).Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error checking column: %v\n", err)
		os.Exit(1)
	}

	if exists {
		fmt.Println("✓ Column 'featured_type' already exists in events table")
		return
	}

	fmt.Println("Column 'featured_type' not found. Adding it now...")

	// Add the column
	addColumnQuery := `
		ALTER TABLE events ADD COLUMN featured_type VARCHAR(10) NOT NULL DEFAULT 'none'
			CHECK (featured_type IN ('none', 'manual', 'auto'))
	`
	
	_, err = db.Exec(addColumnQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error adding column: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully added 'featured_type' column to events table")
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
