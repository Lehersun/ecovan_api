package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// WarehouseService defines the interface for warehouse business logic
type WarehouseService interface {
	// Create creates a new warehouse with validation
	Create(ctx context.Context, req models.CreateWarehouseRequest) (*models.WarehouseResponse, error)

	// GetByID retrieves a warehouse by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.WarehouseResponse, error)

	// List retrieves warehouses with pagination and filtering
	List(ctx context.Context, req models.WarehouseListRequest) (*models.WarehouseListResponse, error)

	// Update updates an existing warehouse with validation
	Update(ctx context.Context, id uuid.UUID, req models.UpdateWarehouseRequest) (*models.WarehouseResponse, error)

	// Delete soft-deletes a warehouse (only if no active equipment)
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted warehouse
	Restore(ctx context.Context, id uuid.UUID) (*models.WarehouseResponse, error)
}
