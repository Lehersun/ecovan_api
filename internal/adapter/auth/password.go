package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters
	iterations = 1
	memory     = 64 * 1024 // 64MB
	threads    = 4
	keyLen     = 32
)

// HashPassword creates a salted hash of the password using Argon2id
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password with Argon2id
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keyLen)

	// Encode the hash and salt in a format that can be stored
	// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%x$%x",
		argon2.Version, memory, iterations, threads, salt, hash)

	return encodedHash, nil
}

// VerifyPassword verifies a password against a stored hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash to extract parameters
	var version int
	var m, t, p uint32
	var salt, hash []byte

	// Parse the hash format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	_, err := fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%x$%x",
		&version, &m, &t, &p, &salt, &hash)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Verify the Argon2 version
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible Argon2 version")
	}

	// Hash the provided password with the same parameters
	computedHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(hash)))

	// Compare the computed hash with the stored hash
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

// IsValidPassword checks if a password meets security requirements
func IsValidPassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more password strength requirements as needed
	// For example: require uppercase, lowercase, numbers, special characters

	return nil
}
