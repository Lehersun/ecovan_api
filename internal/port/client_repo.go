package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientRepository defines the interface for client data operations
type ClientRepository interface {
	// Create creates a new client
	Create(ctx context.Context, client *models.Client) error

	// GetByID retrieves a client by ID, optionally including soft-deleted
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Client, error)

	// List retrieves clients with pagination and filtering
	List(ctx context.Context, req models.ClientListRequest) (*models.ClientListResponse, error)

	// Update updates an existing client
	Update(ctx context.Context, client *models.Client) error

	// SoftDelete marks a client as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted client by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error

	// Exists checks if a client exists with the given name (excluding soft-deleted)
	ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
}
