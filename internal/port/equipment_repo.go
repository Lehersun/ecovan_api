package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// EquipmentRepository defines the interface for equipment data operations
type EquipmentRepository interface {
	// Create creates a new equipment
	Create(ctx context.Context, equipment *models.Equipment) error

	// GetByID retrieves equipment by ID, optionally including soft-deleted
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Equipment, error)

	// List retrieves equipment with pagination and filtering
	List(ctx context.Context, req models.EquipmentListRequest) (*models.EquipmentListResponse, error)

	// Update updates an existing equipment
	Update(ctx context.Context, equipment *models.Equipment) error

	// SoftDelete marks equipment as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted equipment by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error

	// IsAttachedToTransport checks if equipment is currently attached to a transport
	IsAttachedToTransport(ctx context.Context, equipmentID uuid.UUID) (bool, error)

	// ExistsByNumber checks if equipment exists with the given number (excluding soft-deleted)
	ExistsByNumber(ctx context.Context, number string, excludeID *uuid.UUID) (bool, error)
}
