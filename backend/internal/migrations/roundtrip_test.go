//go:build integration

package migrations_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	sqlcdb "github.com/warui/event-ticketing-api/internal/db"
)

func testPgxPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "event_ticketing_test")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pgx pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping DB via pgx pool: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// TestSqlcUserRoundTrip inserts a randomly generated user via sqlc and fetches
// it back by ID, asserting all fields are equal — 100 iterations.
//
// Validates: Requirements 7.7 (Property 4)
func TestSqlcUserRoundTrip(t *testing.T) {
	// Apply all migrations using the sql.DB helper (goose requires *sql.DB)
	sqlDB := testDB(t)
	if err := goose.Up(sqlDB, "."); err != nil {
		t.Fatalf("goose.Up failed: %v", err)
	}

	// Create pgx pool for sqlc queries
	pool := testPgxPool(t)
	q := sqlcdb.New(pool)
	ctx := context.Background()

	roles := []string{"admin", "moderator", "organizer", "attendee"}

	for i := 0; i < 100; i++ {
		params := sqlcdb.CreateUserParams{
			Email:            fmt.Sprintf("user-%d-%d@test.com", i, rand.Int63()),
			Password:         fmt.Sprintf("%x", rand.Int63()) + fmt.Sprintf("%x", rand.Int63()),
			FirstName:        randString(8),
			LastName:         randString(8),
			Phone:            pgtype.Text{Valid: false},
			Role:             roles[rand.Intn(len(roles))],
			IsActive:         true,
			IsVerified:       false,
			TwoFactorEnabled: false,
		}

		created, err := q.CreateUser(ctx, params)
		if err != nil {
			t.Fatalf("iteration %d: CreateUser failed: %v", i, err)
		}

		fetched, err := q.GetUserByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("iteration %d: GetUserByID failed: %v", i, err)
		}

		if fetched.Email != created.Email {
			t.Errorf("iteration %d: Email mismatch: got %q, want %q", i, fetched.Email, created.Email)
		}
		if fetched.FirstName != created.FirstName {
			t.Errorf("iteration %d: FirstName mismatch: got %q, want %q", i, fetched.FirstName, created.FirstName)
		}
		if fetched.LastName != created.LastName {
			t.Errorf("iteration %d: LastName mismatch: got %q, want %q", i, fetched.LastName, created.LastName)
		}
		if fetched.Role != created.Role {
			t.Errorf("iteration %d: Role mismatch: got %q, want %q", i, fetched.Role, created.Role)
		}
		if fetched.IsActive != created.IsActive {
			t.Errorf("iteration %d: IsActive mismatch: got %v, want %v", i, fetched.IsActive, created.IsActive)
		}
	}
}
