package auth

import (
	"strings"
	"testing"
)

const testPassword = "testpassword123"

func TestHashPassword(t *testing.T) {
	password := testPassword

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Check that hash is not empty
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Check that hash starts with expected prefix
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Expected hash to start with $argon2id$, got: %s", hash)
	}

	// Check that hash is different from original password
	if hash == password {
		t.Error("Hash should not equal original password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := testPassword

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify the correct password
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}

	if !valid {
		t.Error("Expected password verification to succeed")
	}

	// Verify incorrect password
	valid, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify wrong password: %v", err)
	}

	if valid {
		t.Error("Expected password verification to fail with wrong password")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	password := testPassword
	invalidHash := "invalid-hash-format"

	// Verify with invalid hash format
	valid, err := VerifyPassword(password, invalidHash)
	if err == nil {
		t.Error("Expected error when parsing invalid hash")
	}

	if valid {
		t.Error("Expected verification to fail with invalid hash")
	}
}

func TestIsValidPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid password", "validpass123", true},
		{"too short", "short", false},
		{"empty password", "", false},
		{"whitespace only", "   ", false},
		{"exactly 8 chars", "12345678", true},
		{"7 chars", "1234567", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := IsValidPassword(tc.password)
			if tc.expected && err != nil {
				t.Errorf("Expected password to be valid, got error: %v", err)
			}
			if !tc.expected && err == nil {
				t.Error("Expected password to be invalid, but no error returned")
			}
		})
	}
}

func TestPasswordHashUniqueness(t *testing.T) {
	password := testPassword

	// Hash the same password multiple times
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password first time: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password second time: %v", err)
	}

	// Hashes should be different due to random salt
	if hash1 == hash2 {
		t.Error("Expected different hashes for the same password due to salt")
	}

	// Both hashes should verify correctly
	valid1, err := VerifyPassword(password, hash1)
	if err != nil {
		t.Fatalf("Failed to verify first hash: %v", err)
	}
	if !valid1 {
		t.Error("First hash should verify correctly")
	}

	valid2, err := VerifyPassword(password, hash2)
	if err != nil {
		t.Fatalf("Failed to verify second hash: %v", err)
	}
	if !valid2 {
		t.Error("Second hash should verify correctly")
	}
}
