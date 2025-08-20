package port

import (
	"context"

	"github.com/google/uuid"
)

// BaseRepository defines common repository operations for entities with soft delete support
type BaseRepository[T any, CreateReq any, UpdateReq any, ListReq any, ListResp any] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity *T) error

	// GetByID retrieves an entity by ID, optionally including soft-deleted
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*T, error)

	// List retrieves entities with pagination and filtering
	List(ctx context.Context, req ListReq) (*ListResp, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity *T) error

	// SoftDelete marks an entity as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted entity by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error
}

// BaseRepositoryWithExists defines common repository operations including existence checks
type BaseRepositoryWithExists[T any, CreateReq any, UpdateReq any, ListReq any, ListResp any] interface {
	BaseRepository[T, CreateReq, UpdateReq, ListReq, ListResp]

	// Exists checks if an entity exists with the given identifier
	Exists(ctx context.Context, identifier interface{}, excludeID *uuid.UUID) (bool, error)
}
