package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// DriverService defines the interface for driver business logic
type DriverService interface {
	// Create creates a new driver with validation
	Create(ctx context.Context, req models.CreateDriverRequest) (*models.DriverResponse, error)

	// GetByID retrieves driver by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.DriverResponse, error)

	// List retrieves drivers with pagination and filtering
	List(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error)

	// Update updates an existing driver with validation
	Update(ctx context.Context, id uuid.UUID, req models.UpdateDriverRequest) (*models.DriverResponse, error)

	// Delete soft-deletes driver (only if not assigned to transport)
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted driver
	Restore(ctx context.Context, id uuid.UUID) (*models.DriverResponse, error)

	// ListAvailable retrieves available drivers (not assigned to any transport)
	ListAvailable(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error)
}
