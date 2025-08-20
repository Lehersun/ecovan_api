package port

import (
	"context"

	"github.com/google/uuid"
)

// BaseService defines common service operations
type BaseService[T any, CreateReq any, UpdateReq any, ListReq any, ListResp any] interface {
	// Create creates a new entity with validation
	Create(ctx context.Context, req CreateReq) (*T, error)

	// GetByID retrieves an entity by ID
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)

	// List retrieves entities with pagination and filtering
	List(ctx context.Context, req ListReq) (*ListResp, error)

	// Update updates an existing entity with validation
	Update(ctx context.Context, id uuid.UUID, req UpdateReq) (*T, error)

	// Delete soft-deletes an entity
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted entity
	Restore(ctx context.Context, id uuid.UUID) (*T, error)
}
