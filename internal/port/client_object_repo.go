package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientObjectRepository defines the interface for client object data operations
type ClientObjectRepository interface {
	BaseRepository[
		models.ClientObject,
		models.ClientObject,
		models.ClientObject,
		models.ClientObjectListRequest,
		models.ClientObjectListResponse,
	]

	// ListByClient retrieves client objects for a specific client
	ListByClient(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error)

	// ExistsByName checks if a client object with the given name exists for a client
	ExistsByName(ctx context.Context, clientID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)

	// ExistsByAddress checks if a client object exists with the given address for a client
	ExistsByAddress(ctx context.Context, clientID uuid.UUID, address string, excludeID *uuid.UUID) (bool, error)

	// GetDeleteConflicts returns detailed information about what prevents deletion
	GetDeleteConflicts(ctx context.Context, clientObjectID uuid.UUID) (*models.DeleteConflicts, error)
}
