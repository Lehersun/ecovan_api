package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// EquipmentService defines the interface for equipment business logic
type EquipmentService interface {
	// Create creates a new equipment with validation
	Create(ctx context.Context, req models.CreateEquipmentRequest) (*models.EquipmentResponse, error)

	// GetByID retrieves equipment by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.EquipmentResponse, error)

	// List retrieves equipment with pagination and filtering
	List(ctx context.Context, req models.EquipmentListRequest) (*models.EquipmentListResponse, error)

	// Update updates an existing equipment with validation
	Update(ctx context.Context, id uuid.UUID, req models.UpdateEquipmentRequest) (*models.EquipmentResponse, error)

	// Delete soft-deletes equipment (only if not attached to transport)
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted equipment
	Restore(ctx context.Context, id uuid.UUID) (*models.EquipmentResponse, error)
}
