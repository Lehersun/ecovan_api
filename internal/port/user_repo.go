package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user and returns the created user
	Create(ctx context.Context, email, passwordHash, role string) (*models.User, error)

	// Get retrieves a user by ID
	Get(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// List retrieves a paginated list of users
	List(ctx context.Context, page, pageSize int) ([]*models.User, int, error)

	// Delete removes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error)

	// FindByEmail is an alias for GetByEmail for backward compatibility
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}
