package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// DriverRepository defines the interface for driver data operations
type DriverRepository interface {
	BaseRepository[models.Driver, models.Driver, models.Driver, models.DriverListRequest, models.DriverListResponse]

	// ListAvailable retrieves available drivers (not assigned to any transport)
	ListAvailable(ctx context.Context) ([]models.Driver, error)

	// ExistsByLicenseNo checks if driver exists with the given license number
	ExistsByLicenseNo(ctx context.Context, licenseNo string, excludeID *uuid.UUID) (bool, error)

	// IsAssignedToTransport checks if driver is currently assigned to a transport
	IsAssignedToTransport(ctx context.Context, driverID uuid.UUID) (bool, error)
}
