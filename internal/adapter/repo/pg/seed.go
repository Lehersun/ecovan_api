package pg

import (
	"context"
	"fmt"
	"os"

	"eco-van-api/internal/adapter/auth"
	"eco-van-api/internal/models"
)

// SeedAdminUser creates an admin user if it doesn't exist
func (r *UserRepository) SeedAdminUser(ctx context.Context) error {
	// Check if admin user already exists
	existingUser, err := r.FindByEmail(ctx, "admin@example.com")
	if err == nil && existingUser != nil {
		// Admin user already exists
		return nil
	}

	// Get admin password from environment
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		// Use default password if not set
		adminPassword = "admin123456"
	}

	// Validate password
	if err := auth.IsValidPassword(adminPassword); err != nil {
		return fmt.Errorf("invalid admin password: %w", err)
	}

	// Hash password
	passwordHash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	_, err = r.Create(ctx, "admin@example.com", passwordHash, models.UserRoleAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Admin user created successfully
	return nil
}
