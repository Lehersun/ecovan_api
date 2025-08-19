package port

import (
	"context"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientObjectService defines the interface for client object business logic
type ClientObjectService interface {
	// Create creates a new client object for a client
	Create(ctx context.Context, clientID uuid.UUID, req models.CreateClientObjectRequest) (*models.ClientObjectResponse, error)

	// GetByID retrieves a client object by ID
	GetByID(ctx context.Context, clientID, id uuid.UUID, includeDeleted bool) (*models.ClientObjectResponse, error)

	// List retrieves client objects for a specific client with pagination
	List(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error)

	// Update updates an existing client object
	Update(ctx context.Context, clientID, id uuid.UUID, req models.UpdateClientObjectRequest) (*models.ClientObjectResponse, error)

	// Delete soft deletes a client object (guarded)
	Delete(ctx context.Context, clientID, id uuid.UUID) error

	// Restore restores a soft-deleted client object
	Restore(ctx context.Context, clientID, id uuid.UUID) (*models.ClientObjectResponse, error)
}
