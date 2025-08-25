package service

import (
	"context"
	"fmt"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
)

// orderService implements port.OrderService
type orderService struct {
	orderRepo     port.OrderRepository
	clientRepo    port.ClientRepository
	clientObjRepo port.ClientObjectRepository
	transportRepo port.TransportRepository
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo port.OrderRepository,
	clientRepo port.ClientRepository,
	clientObjRepo port.ClientObjectRepository,
	transportRepo port.TransportRepository,
) port.OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		clientRepo:    clientRepo,
		clientObjRepo: clientObjRepo,
		transportRepo: transportRepo,
	}
}

// Create creates a new order with validation
func (s *orderService) Create(ctx context.Context, req *models.CreateOrderRequest, createdBy *uuid.UUID) (*models.OrderResponse, error) {
	// Validate that client exists
	client, err := s.clientRepo.GetByID(ctx, req.ClientID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	// Validate that client object exists and belongs to the client
	clientObj, err := s.clientObjRepo.GetByID(ctx, req.ObjectID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get client object: %w", err)
	}
	if clientObj == nil {
		return nil, fmt.Errorf("client object not found")
	}
	if clientObj.ClientID != req.ClientID {
		return nil, fmt.Errorf("client object does not belong to the specified client")
	}

	// Validate transport if being assigned
	if req.TransportID != nil {
		transport, err := s.transportRepo.GetByID(ctx, *req.TransportID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get transport: %w", err)
		}
		if transport == nil {
			return nil, fmt.Errorf("transport not found")
		}
		// Check if transport is available (has IN_WORK status)
		if transport.Status != "IN_WORK" {
			return nil, fmt.Errorf("transport is not available (status: %s)", transport.Status)
		}
	}

	// Create order from request
	order := models.FromOrderCreateRequest(req)
	order.CreatedBy = createdBy

	// Save to repository
	err = s.orderRepo.Create(ctx, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Return response
	response := order.ToResponse()
	return &response, nil
}

// GetByID retrieves an order by ID
func (s *orderService) GetByID(ctx context.Context, id uuid.UUID) (*models.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	response := order.ToResponse()
	return &response, nil
}

// List retrieves orders with pagination and filtering
func (s *orderService) List(ctx context.Context, req models.OrderListRequest) (*models.OrderListResponse, error) {
	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	const maxPageSize = 100
	if req.PageSize > maxPageSize {
		req.PageSize = maxPageSize
	}

	response, err := s.orderRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return response, nil
}

// validateClientUpdate validates client update if provided
func (s *orderService) validateClientUpdate(ctx context.Context, clientID *uuid.UUID) error {
	if clientID == nil {
		return nil
	}

	client, err := s.clientRepo.GetByID(ctx, *clientID, false)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	if client == nil {
		return fmt.Errorf("client not found")
	}
	return nil
}

// clientObjectValidationParams holds parameters for client object validation
type clientObjectValidationParams struct {
	objectID         *uuid.UUID
	clientID         *uuid.UUID
	existingClientID uuid.UUID
}

// validateClientObjectUpdate validates client object update if provided
func (s *orderService) validateClientObjectUpdate(ctx context.Context, params clientObjectValidationParams) error {
	if params.objectID == nil {
		return nil
	}

	clientObj, err := s.clientObjRepo.GetByID(ctx, *params.objectID, false)
	if err != nil {
		return fmt.Errorf("failed to get client object: %w", err)
	}
	if clientObj == nil {
		return fmt.Errorf("client object not found")
	}

	// If client is also being updated, validate the relationship
	if params.clientID != nil {
		if clientObj.ClientID != *params.clientID {
			return fmt.Errorf("client object does not belong to the specified client")
		}
	} else {
		// Use existing client ID for validation
		if clientObj.ClientID != params.existingClientID {
			return fmt.Errorf("client object does not belong to the order's client")
		}
	}
	return nil
}

// validateTransportUpdate validates transport update if provided
func (s *orderService) validateTransportUpdate(ctx context.Context, transportID *uuid.UUID) error {
	if transportID == nil {
		return nil
	}

	transport, err := s.transportRepo.GetByID(ctx, *transportID, false)
	if err != nil {
		return fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return fmt.Errorf("transport not found")
	}
	// Check if transport is available (has IN_WORK status)
	if transport.Status != "IN_WORK" {
		return fmt.Errorf("transport is not available (status: %s)", transport.Status)
	}
	return nil
}

// Update updates an existing order with validation
func (s *orderService) Update(ctx context.Context, id uuid.UUID, req models.UpdateOrderRequest) (*models.OrderResponse, error) {
	// Get existing order
	order, err := s.orderRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	// Validate updates
	if err := s.validateClientUpdate(ctx, req.ClientID); err != nil {
		return nil, err
	}

	if err := s.validateClientObjectUpdate(ctx, clientObjectValidationParams{
		objectID:         req.ObjectID,
		clientID:         req.ClientID,
		existingClientID: order.ClientID,
	}); err != nil {
		return nil, err
	}

	if err := s.validateTransportUpdate(ctx, req.TransportID); err != nil {
		return nil, err
	}

	// Update order from request
	order.UpdateFromRequest(req)

	// Save to repository
	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Return updated response
	response := order.ToResponse()
	return &response, nil
}

// UpdateStatus updates the order status with transition validation
func (s *orderService) UpdateStatus(ctx context.Context, id uuid.UUID, req models.UpdateOrderStatusRequest) (*models.OrderResponse, error) {
	// Get existing order
	order, err := s.orderRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	// Validate status transition
	err = order.CanTransitionTo(req.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	// Update status
	order.Status = string(req.Status)

	// Save to repository
	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Return updated response
	response := order.ToResponse()
	return &response, nil
}

// Delete soft-deletes an order (only if status allows)
func (s *orderService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get existing order
	order, err := s.orderRepo.GetByID(ctx, id, false)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Check if order can be deleted
	err = order.CanBeDeleted()
	if err != nil {
		return fmt.Errorf("order cannot be deleted: %w", err)
	}

	// Soft delete the order
	err = s.orderRepo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted order
func (s *orderService) Restore(ctx context.Context, id uuid.UUID) (*models.OrderResponse, error) {
	// Check if order exists and is deleted
	order, err := s.orderRepo.GetByID(ctx, id, true) // Include soft-deleted
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	if order.DeletedAt == nil {
		return nil, fmt.Errorf("order is not deleted")
	}

	// Restore the order
	err = s.orderRepo.Restore(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to restore order: %w", err)
	}

	// Get the restored order
	restoredOrder, err := s.orderRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored order: %w", err)
	}

	// Return restored response
	response := restoredOrder.ToResponse()
	return &response, nil
}

// AssignTransport assigns transport to an order
func (s *orderService) AssignTransport(ctx context.Context, orderID uuid.UUID, req models.AssignTransportRequest) error {
	// Check if order exists and is not deleted
	order, err := s.orderRepo.GetByID(ctx, orderID, false)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Check if transport exists and is not deleted
	transport, err := s.transportRepo.GetByID(ctx, req.TransportID, false)
	if err != nil {
		return fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return fmt.Errorf("transport not found")
	}

	// Assign transport to order
	order.AssignTransport(req.TransportID)

	// Save to repository
	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to assign transport to order: %w", err)
	}

	return nil
}
