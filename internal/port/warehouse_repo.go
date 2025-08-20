package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// WarehouseRepository defines the interface for warehouse data operations
type WarehouseRepository interface {
	BaseRepository[models.Warehouse, models.Warehouse, models.Warehouse, models.WarehouseListRequest, models.WarehouseListResponse]

	// ExistsByName checks if a warehouse exists with the given name (excluding soft-deleted)
	ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)

	// HasActiveEquipment checks if a warehouse has any non-deleted equipment
	HasActiveEquipment(ctx context.Context, warehouseID uuid.UUID) (bool, error)
}
