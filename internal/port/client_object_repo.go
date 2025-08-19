package port

import (
	"context"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientObjectRepository defines the interface for client object data access
type ClientObjectRepository interface {
	// Create creates a new client object
	Create(ctx context.Context, clientObject *models.ClientObject) error

	// GetByID retrieves a client object by ID
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.ClientObject, error)

	// List retrieves client objects for a specific client with pagination
	List(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error)

	// Update updates an existing client object
	Update(ctx context.Context, clientObject *models.ClientObject) error

	// SoftDelete soft deletes a client object (guarded)
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted client object
	Restore(ctx context.Context, id uuid.UUID) error

	// ExistsByName checks if a client object with the given name exists for a client
	ExistsByName(ctx context.Context, clientID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)

	// HasActiveOrders checks if there are active orders for this client object
	HasActiveOrders(ctx context.Context, clientObjectID uuid.UUID) (bool, error)

	// HasActiveEquipment checks if there is active equipment placed at this client object
	HasActiveEquipment(ctx context.Context, clientObjectID uuid.UUID) (bool, error)

	// GetDeleteConflicts returns detailed information about what prevents deletion
	GetDeleteConflicts(ctx context.Context, clientObjectID uuid.UUID) (*models.DeleteConflicts, error)
}
