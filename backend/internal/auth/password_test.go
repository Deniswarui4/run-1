package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "TestPassword123"
	wrongPassword := "WrongPassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	// Test wrong password
	if CheckPassword(wrongPassword, hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	password := "TestPassword123"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Hashes should be different due to salt
	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different")
	}

	// But both should validate correctly
	if !CheckPassword(password, hash1) {
		t.Error("First hash should validate")
	}
	if !CheckPassword(password, hash2) {
		t.Error("Second hash should validate")
	}
}
