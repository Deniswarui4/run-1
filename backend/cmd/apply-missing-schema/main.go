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

	fmt.Println("Applying missing schema changes from migration 00003...")
	fmt.Println()

	// Apply all the missing schema changes
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "Create guest_sessions table",
			sql: `
				CREATE TABLE IF NOT EXISTS guest_sessions (
					id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					token      VARCHAR NOT NULL UNIQUE,
					email      VARCHAR NOT NULL,
					expires_at TIMESTAMPTZ NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_guest_sessions_token ON guest_sessions (token);
				CREATE INDEX IF NOT EXISTS idx_guest_sessions_expires_at ON guest_sessions (expires_at);
			`,
		},
		{
			name: "Add guest_email to tickets",
			sql: `
				DO $$ 
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.columns 
						WHERE table_name = 'tickets' AND column_name = 'guest_email'
					) THEN
						ALTER TABLE tickets ADD COLUMN guest_email VARCHAR;
						ALTER TABLE tickets ALTER COLUMN attendee_id DROP NOT NULL;
					END IF;
				END $$;
			`,
		},
		{
			name: "Create draft_carts table",
			sql: `
				CREATE TABLE IF NOT EXISTS draft_carts (
					id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id         UUID REFERENCES users(id),
					guest_session   VARCHAR,
					event_id        UUID NOT NULL REFERENCES events(id),
					status          VARCHAR(20) NOT NULL DEFAULT 'active'
										CHECK (status IN ('active', 'expired', 'completed')),
					expires_at      TIMESTAMPTZ NOT NULL,
					created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_draft_carts_user_id ON draft_carts (user_id);
				CREATE INDEX IF NOT EXISTS idx_draft_carts_guest_session ON draft_carts (guest_session);
				CREATE INDEX IF NOT EXISTS idx_draft_carts_expires_at ON draft_carts (expires_at);
			`,
		},
		{
			name: "Create draft_cart_items table",
			sql: `
				CREATE TABLE IF NOT EXISTS draft_cart_items (
					id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					cart_id        UUID NOT NULL REFERENCES draft_carts(id) ON DELETE CASCADE,
					ticket_type_id UUID NOT NULL REFERENCES ticket_types(id),
					quantity       INTEGER NOT NULL CHECK (quantity > 0),
					created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					UNIQUE (cart_id, ticket_type_id)
				);
				CREATE INDEX IF NOT EXISTS idx_draft_cart_items_cart_id ON draft_cart_items (cart_id);
			`,
		},
		{
			name: "Add auto_feature_threshold to platform_settings",
			sql: `
				DO $$ 
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.columns 
						WHERE table_name = 'platform_settings' AND column_name = 'auto_feature_threshold'
					) THEN
						ALTER TABLE platform_settings ADD COLUMN auto_feature_threshold INTEGER NOT NULL DEFAULT 0;
					END IF;
				END $$;
			`,
		},
		{
			name: "Create event_feature_evaluations table",
			sql: `
				CREATE TABLE IF NOT EXISTS event_feature_evaluations (
					event_id                    UUID PRIMARY KEY REFERENCES events(id),
					consecutive_below_threshold INTEGER NOT NULL DEFAULT 0,
					last_evaluated_at           TIMESTAMPTZ
				);
			`,
		},
	}

	for _, migration := range migrations {
		fmt.Printf("Applying: %s...\n", migration.name)
		_, err := db.Exec(migration.sql)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Success\n\n")
	}

	fmt.Println("✓ All missing schema changes have been applied successfully!")
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
