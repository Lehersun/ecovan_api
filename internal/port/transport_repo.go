package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// TransportRepository defines the interface for transport data operations
type TransportRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, transport *models.Transport) error
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Transport, error)
	Update(ctx context.Context, transport *models.Transport) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error)

	// ExistsByPlateNo checks if a transport exists with the given plate number (excluding soft-deleted)
	ExistsByPlateNo(ctx context.Context, plateNo string, excludeID *uuid.UUID) (bool, error)

	// HasActiveDriver checks if transport has an assigned driver
	HasActiveDriver(ctx context.Context, transportID uuid.UUID) (bool, error)

	// HasActiveEquipment checks if transport has assigned equipment
	HasActiveEquipment(ctx context.Context, transportID uuid.UUID) (bool, error)

	// HasActiveOrders checks if transport has active orders (DRAFT, SCHEDULED, IN_PROGRESS)
	HasActiveOrders(ctx context.Context, transportID uuid.UUID) (bool, error)

	// AssignDriver assigns a driver to transport
	AssignDriver(ctx context.Context, transportID, driverID uuid.UUID) error

	// UnassignDriver removes driver assignment from transport
	UnassignDriver(ctx context.Context, transportID uuid.UUID) error

	// AssignEquipment assigns equipment to transport
	AssignEquipment(ctx context.Context, transportID, equipmentID uuid.UUID) error

	// UnassignEquipment removes equipment assignment from transport
	UnassignEquipment(ctx context.Context, transportID uuid.UUID) error

	// GetAvailable returns transport with status IN_WORK and no soft-delete
	GetAvailable(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error)

	// IsDriverAssignedToOtherTransport checks if driver is assigned to another non-deleted transport
	IsDriverAssignedToOtherTransport(ctx context.Context, driverID, excludeTransportID uuid.UUID) (bool, error)

	// IsEquipmentAssignedToOtherTransport checks if equipment is assigned to another non-deleted transport
	IsEquipmentAssignedToOtherTransport(ctx context.Context, equipmentID, excludeTransportID uuid.UUID) (bool, error)

	// IsEquipmentAvailableForAssignment checks if equipment can be assigned (no client_object_id or warehouse_id)
	IsEquipmentAvailableForAssignment(ctx context.Context, equipmentID uuid.UUID) (bool, error)
}
