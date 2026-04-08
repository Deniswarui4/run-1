package database

import (
	"testing"

	"github.com/warui/event-ticketing-api/internal/config"
)

func TestInitDB(t *testing.T) {
	// Skip if no test database is available
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "event_ticketing_test",
		DBSSLMode:  "disable",
	}

	// This test will only pass if a test database is available
	// In CI/CD, you would set up a test database
	t.Run("Database connection", func(t *testing.T) {
		db, err := InitDB(cfg)
		if err != nil {
			t.Skipf("Skipping test: database not available: %v", err)
			return
		}

		if db == nil {
			t.Error("Database should not be nil")
		}

		// Test connection
		sqlDB, err := db.DB()
		if err != nil {
			t.Errorf("Failed to get database instance: %v", err)
		}

		if err := sqlDB.Ping(); err != nil {
			t.Errorf("Failed to ping database: %v", err)
		}
	})
}


