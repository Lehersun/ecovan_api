package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// DriverRepository defines the interface for driver data operations
type DriverRepository interface {
	// Create creates a new driver
	Create(ctx context.Context, driver *models.Driver) error

	// GetByID retrieves driver by ID, optionally including soft-deleted
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Driver, error)

	// List retrieves drivers with pagination and filtering
	List(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error)

	// Update updates an existing driver
	Update(ctx context.Context, driver *models.Driver) error

	// SoftDelete marks driver as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted driver by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error

	// IsAssignedToTransport checks if driver is currently assigned to a transport
	IsAssignedToTransport(ctx context.Context, driverID uuid.UUID) (bool, error)

	// ExistsByLicenseNo checks if driver exists with the given license number (excluding soft-deleted)
	ExistsByLicenseNo(ctx context.Context, licenseNo string, excludeID *uuid.UUID) (bool, error)

	// ListAvailable retrieves available drivers (not assigned to any transport)
	ListAvailable(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error)
}
