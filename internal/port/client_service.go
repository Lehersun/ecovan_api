package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientService defines the interface for client business logic
type ClientService interface {
	// Create creates a new client with validation
	Create(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)

	// GetByID retrieves a client by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.ClientResponse, error)

	// List retrieves clients with pagination and filtering
	List(ctx context.Context, req models.ClientListRequest) (*models.ClientListResponse, error)

	// Update updates an existing client with validation
	Update(ctx context.Context, id uuid.UUID, req models.UpdateClientRequest) (*models.ClientResponse, error)

	// Delete soft-deletes a client
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted client
	Restore(ctx context.Context, id uuid.UUID) (*models.ClientResponse, error)
}
