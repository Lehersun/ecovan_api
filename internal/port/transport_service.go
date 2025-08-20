package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// TransportService defines the interface for transport business logic
type TransportService interface {
	// Basic CRUD operations
	Create(ctx context.Context, req models.CreateTransportRequest) (*models.TransportResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.TransportResponse, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdateTransportRequest) (*models.TransportResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) (*models.TransportResponse, error)
	List(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error)

	// GetAvailable returns available transport (IN_WORK status, non-deleted)
	GetAvailable(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error)

	// AssignDriver assigns a driver to transport with validation
	AssignDriver(ctx context.Context, transportID string, req models.AssignDriverRequest) error

	// AssignEquipment assigns equipment to transport with validation
	AssignEquipment(ctx context.Context, transportID string, req models.AssignEquipmentRequest) error
}
