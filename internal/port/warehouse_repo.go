package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// WarehouseRepository defines the interface for warehouse data operations
type WarehouseRepository interface {
	// Create creates a new warehouse
	Create(ctx context.Context, warehouse *models.Warehouse) error

	// GetByID retrieves a warehouse by ID, optionally including soft-deleted
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Warehouse, error)

	// List retrieves warehouses with pagination and filtering
	List(ctx context.Context, req models.WarehouseListRequest) (*models.WarehouseListResponse, error)

	// Update updates an existing warehouse
	Update(ctx context.Context, warehouse *models.Warehouse) error

	// SoftDelete marks a warehouse as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted warehouse by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error

	// Exists checks if a warehouse exists with the given name (excluding soft-deleted)
	ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)

	// HasActiveEquipment checks if a warehouse has any non-deleted equipment
	HasActiveEquipment(ctx context.Context, warehouseID uuid.UUID) (bool, error)
}
