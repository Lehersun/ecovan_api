package port

import (
	"context"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// OrderService defines the interface for order business logic operations
type OrderService interface {
	// Create creates a new order with validation
	Create(ctx context.Context, req models.CreateOrderRequest, createdBy *uuid.UUID) (*models.OrderResponse, error)

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.OrderResponse, error)

	// List retrieves orders with pagination and filtering
	List(ctx context.Context, req models.OrderListRequest) (*models.OrderListResponse, error)

	// Update updates an existing order with validation
	Update(ctx context.Context, id uuid.UUID, req models.UpdateOrderRequest) (*models.OrderResponse, error)

	// UpdateStatus updates the order status with transition validation
	UpdateStatus(ctx context.Context, id uuid.UUID, req models.UpdateOrderStatusRequest) (*models.OrderResponse, error)

	// Delete soft-deletes an order (only if status allows)
	Delete(ctx context.Context, id uuid.UUID) error

	// Restore restores a soft-deleted order
	Restore(ctx context.Context, id uuid.UUID) (*models.OrderResponse, error)

	// AssignTransport assigns transport to an order
	AssignTransport(ctx context.Context, orderID uuid.UUID, req models.AssignTransportRequest) error
}
