package port

import (
	"context"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// ClientRepository defines the interface for client data operations
type ClientRepository interface {
	BaseRepository[models.Client, models.Client, models.Client, models.ClientListRequest, models.ClientListResponse]

	// ExistsByName checks if a client exists with the given name (excluding soft-deleted)
	ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
}
