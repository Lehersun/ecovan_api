package port

import (
	"context"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// OrderRepository defines the interface for order data access operations
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *models.Order) error

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *models.Order) error

	// SoftDelete marks an order as deleted by setting deleted_at
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted order by clearing deleted_at
	Restore(ctx context.Context, id uuid.UUID) error

	// List retrieves orders with pagination and filtering
	List(ctx context.Context, req models.OrderListRequest) (*models.OrderListResponse, error)

	// ExistsByClientAndObject checks if an order exists for the given client and object
	ExistsByClientAndObject(ctx context.Context, clientID, objectID uuid.UUID, excludeID *uuid.UUID) (bool, error)

	// HasActiveOrders checks if a client object has any active orders
	HasActiveOrders(ctx context.Context, objectID uuid.UUID) (bool, error)

	// GetActiveOrdersByObject returns active orders for a specific object
	GetActiveOrdersByObject(ctx context.Context, objectID uuid.UUID) ([]models.Order, error)
}
