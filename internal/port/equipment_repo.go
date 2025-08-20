package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// EquipmentRepository defines the interface for equipment data operations
type EquipmentRepository interface {
	BaseRepository[models.Equipment, models.Equipment, models.Equipment, models.EquipmentListRequest, models.EquipmentListResponse]

	// ExistsByNumber checks if equipment exists with the given number (excluding soft-deleted)
	ExistsByNumber(ctx context.Context, number string, excludeID *uuid.UUID) (bool, error)

	// IsAttachedToTransport checks if equipment is currently attached to a transport
	IsAttachedToTransport(ctx context.Context, equipmentID uuid.UUID) (bool, error)
}
