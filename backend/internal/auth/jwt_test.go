package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
)

func TestGenerateToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	secret := "test-secret-key"
	expiryHours := 24

	token, err := GenerateToken(user, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestValidateToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	secret := "test-secret-key"
	expiryHours := 24

	token, err := GenerateToken(user, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %v, got %v", user.ID, claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, claims.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected Role %s, got %s", user.Role, claims.Role)
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"
	expiryHours := 24

	token, err := GenerateToken(user, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, wrongSecret)
	if err == nil {
		t.Error("ValidateToken should fail with wrong secret")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	secret := "test-secret-key"
	expiryHours := -1 // Expired token

	token, err := GenerateToken(user, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Wait a moment to ensure token is expired
	time.Sleep(100 * time.Millisecond)

	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("ValidateToken should fail for expired token")
	}
}

func TestRefreshToken(t *testing.T) {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	secret := "test-secret-key"
	expiryHours := 24

	originalToken, err := GenerateToken(user, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	newToken, err := RefreshToken(originalToken, secret, expiryHours)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	if newToken == "" {
		t.Error("Refreshed token should not be empty")
	}

	// Note: Tokens might be the same if generated in the same second
	// This is expected behavior and not an error

	// Validate the new token
	claims, err := ValidateToken(newToken, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed for refreshed token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Error("Refreshed token should have same user ID")
	}
}
