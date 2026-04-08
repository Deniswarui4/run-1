package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set some test environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "test-secret")

	cfg := LoadConfig()

	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected Port 9090, got %s", cfg.Port)
	}

	if cfg.DBHost != "testhost" {
		t.Errorf("Expected DBHost testhost, got %s", cfg.DBHost)
	}

	if cfg.DBName != "testdb" {
		t.Errorf("Expected DBName testdb, got %s", cfg.DBName)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("Expected JWTSecret test-secret, got %s", cfg.JWTSecret)
	}

	// Clean up
	os.Unsetenv("PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("JWT_SECRET")
}

func TestLoadConfigDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Clearenv()

	cfg := LoadConfig()

	if cfg.Port != "8080" {
		t.Errorf("Expected default Port 8080, got %s", cfg.Port)
	}

	if cfg.GinMode != "debug" {
		t.Errorf("Expected default GinMode debug, got %s", cfg.GinMode)
	}

	if cfg.DBPort != "5432" {
		t.Errorf("Expected default DBPort 5432, got %s", cfg.DBPort)
	}

	if cfg.Currency != "NGN" {
		t.Errorf("Expected default Currency NGN, got %s", cfg.Currency)
	}

	if cfg.RateLimitWindow != time.Minute {
		t.Errorf("Expected default RateLimitWindow 1m, got %v", cfg.RateLimitWindow)
	}
}

func TestLoadConfigRateLimiting(t *testing.T) {
	os.Setenv("RATE_LIMIT_REQUESTS", "200")
	os.Setenv("RATE_LIMIT_WINDOW", "2m")

	cfg := LoadConfig()

	if cfg.RateLimitRequests != 200 {
		t.Errorf("Expected RateLimitRequests 200, got %d", cfg.RateLimitRequests)
	}

	if cfg.RateLimitWindow != 2*time.Minute {
		t.Errorf("Expected RateLimitWindow 2m, got %v", cfg.RateLimitWindow)
	}

	os.Unsetenv("RATE_LIMIT_REQUESTS")
	os.Unsetenv("RATE_LIMIT_WINDOW")
}

func TestLoadConfigPlatformFees(t *testing.T) {
	os.Setenv("DEFAULT_PLATFORM_FEE_PERCENTAGE", "7.5")
	os.Setenv("DEFAULT_WITHDRAWAL_FEE_PERCENTAGE", "3.0")

	cfg := LoadConfig()

	if cfg.DefaultPlatformFeePercentage != 7.5 {
		t.Errorf("Expected DefaultPlatformFeePercentage 7.5, got %f", cfg.DefaultPlatformFeePercentage)
	}

	if cfg.DefaultWithdrawalFeePercentage != 3.0 {
		t.Errorf("Expected DefaultWithdrawalFeePercentage 3.0, got %f", cfg.DefaultWithdrawalFeePercentage)
	}

	os.Unsetenv("DEFAULT_PLATFORM_FEE_PERCENTAGE")
	os.Unsetenv("DEFAULT_WITHDRAWAL_FEE_PERCENTAGE")
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected test_value, got %s", result)
	}

	result = getEnv("NON_EXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected default, got %s", result)
	}

	os.Unsetenv("TEST_VAR")
}
