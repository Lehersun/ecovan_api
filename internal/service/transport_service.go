package service

import (
	"context"
	"fmt"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
)

// TransportService implements port.TransportService
type TransportService struct {
	transportRepo port.TransportRepository
	driverRepo    port.DriverRepository
	equipmentRepo port.EquipmentRepository
}

// NewTransportService creates a new TransportService
func NewTransportService(
	transportRepo port.TransportRepository,
	driverRepo port.DriverRepository,
	equipmentRepo port.EquipmentRepository,
) port.TransportService {
	return &TransportService{
		transportRepo: transportRepo,
		driverRepo:    driverRepo,
		equipmentRepo: equipmentRepo,
	}
}

// Create creates a new transport
//
//nolint:dupl // Similar pattern across services but with different business logic
func (s *TransportService) Create(ctx context.Context, req models.CreateTransportRequest) (*models.TransportResponse, error) {
	// Check if plate number already exists
	exists, err := s.transportRepo.ExistsByPlateNo(ctx, req.PlateNo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check plate number uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("transport with plate number %s already exists", req.PlateNo)
	}

	// Create transport from request
	transport := models.FromTransportCreateRequest(req)

	// Save to repository
	err = s.transportRepo.Create(ctx, &transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Return response
	response := transport.ToResponse()
	return &response, nil
}

// GetByID retrieves a transport by ID
func (s *TransportService) GetByID(ctx context.Context, id uuid.UUID) (*models.TransportResponse, error) {
	transport, err := s.transportRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return nil, fmt.Errorf("transport not found")
	}

	response := transport.ToResponse()
	return &response, nil
}

// Update updates an existing transport
func (s *TransportService) Update(ctx context.Context, id uuid.UUID, req models.UpdateTransportRequest) (*models.TransportResponse, error) {
	// Get existing transport
	transport, err := s.transportRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return nil, fmt.Errorf("transport not found")
	}

	// Check plate number uniqueness if updating
	if req.PlateNo != nil && *req.PlateNo != transport.PlateNo {
		exists, err := s.transportRepo.ExistsByPlateNo(ctx, *req.PlateNo, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check plate number uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("transport with plate number %s already exists", *req.PlateNo)
		}
	}

	// Update transport from request
	transport.UpdateFromRequest(req)

	// Save to repository
	err = s.transportRepo.Update(ctx, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to update transport: %w", err)
	}

	// Return response
	response := transport.ToResponse()
	return &response, nil
}

// Delete soft-deletes a transport
func (s *TransportService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if transport has active driver
	hasDriver, err := s.transportRepo.HasActiveDriver(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check driver assignment: %w", err)
	}
	if hasDriver {
		return fmt.Errorf("cannot delete transport: driver is currently assigned")
	}

	// Check if transport has active equipment
	hasEquipment, err := s.transportRepo.HasActiveEquipment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check equipment assignment: %w", err)
	}
	if hasEquipment {
		return fmt.Errorf("cannot delete transport: equipment is currently assigned")
	}

	// Check if transport has active orders
	hasOrders, err := s.transportRepo.HasActiveOrders(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check active orders: %w", err)
	}
	if hasOrders {
		return fmt.Errorf("cannot delete transport: has active orders")
	}

	// Soft delete transport
	err = s.transportRepo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete transport: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted transport
func (s *TransportService) Restore(ctx context.Context, id uuid.UUID) (*models.TransportResponse, error) {
	// Restore transport
	err := s.transportRepo.Restore(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to restore transport: %w", err)
	}

	// Get restored transport
	transport, err := s.transportRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored transport: %w", err)
	}

	response := transport.ToResponse()
	return &response, nil
}

// List lists transport with filtering and pagination
func (s *TransportService) List(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error) {
	response, err := s.transportRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list transport: %w", err)
	}

	return response, nil
}

// GetAvailable returns available transport (IN_WORK status, non-deleted)
func (s *TransportService) GetAvailable(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error) {
	response, err := s.transportRepo.GetAvailable(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get available transport: %w", err)
	}

	return response, nil
}

// AssignDriver assigns a driver to transport with validation
func (s *TransportService) AssignDriver(ctx context.Context, transportID string, req models.AssignDriverRequest) error {
	// Parse transport ID
	tID, err := uuid.Parse(transportID)
	if err != nil {
		return fmt.Errorf("invalid transport ID: %w", err)
	}

	// Check if transport exists and is not deleted
	transport, err := s.transportRepo.GetByID(ctx, tID, false)
	if err != nil {
		return fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return fmt.Errorf("transport not found")
	}

	// Check if driver exists and is not deleted
	driver, err := s.driverRepo.GetByID(ctx, req.DriverID, false)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}
	if driver == nil {
		return fmt.Errorf("driver not found")
	}

	// Check if driver is already assigned to another transport
	isAssigned, err := s.transportRepo.IsDriverAssignedToOtherTransport(ctx, req.DriverID, tID)
	if err != nil {
		return fmt.Errorf("failed to check driver assignment: %w", err)
	}
	if isAssigned {
		return fmt.Errorf("driver is already assigned to another transport")
	}

	// Assign driver to transport
	err = s.transportRepo.AssignDriver(ctx, tID, req.DriverID)
	if err != nil {
		return fmt.Errorf("failed to assign driver: %w", err)
	}

	return nil
}

// AssignEquipment assigns equipment to transport with validation
func (s *TransportService) AssignEquipment(ctx context.Context, transportID string, req models.AssignEquipmentRequest) error {
	// Parse transport ID
	tID, err := uuid.Parse(transportID)
	if err != nil {
		return fmt.Errorf("invalid transport ID: %w", err)
	}

	// Check if transport exists and is not deleted
	transport, err := s.transportRepo.GetByID(ctx, tID, false)
	if err != nil {
		return fmt.Errorf("failed to get transport: %w", err)
	}
	if transport == nil {
		return fmt.Errorf("transport not found")
	}

	// Check if equipment exists and is not deleted
	equipment, err := s.equipmentRepo.GetByID(ctx, req.EquipmentID, false)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}

	// Check if equipment is available for assignment (no client_object_id or warehouse_id)
	isAvailable, err := s.transportRepo.IsEquipmentAvailableForAssignment(ctx, req.EquipmentID)
	if err != nil {
		return fmt.Errorf("failed to check equipment availability: %w", err)
	}
	if !isAvailable {
		return fmt.Errorf("equipment is not available for assignment (already placed at client object or warehouse)")
	}

	// Check if equipment is already assigned to another transport
	isAssigned, err := s.transportRepo.IsEquipmentAssignedToOtherTransport(ctx, req.EquipmentID, tID)
	if err != nil {
		return fmt.Errorf("failed to check equipment assignment: %w", err)
	}
	if isAssigned {
		return fmt.Errorf("equipment is already assigned to another transport")
	}

	// Assign equipment to transport
	err = s.transportRepo.AssignEquipment(ctx, tID, req.EquipmentID)
	if err != nil {
		return fmt.Errorf("failed to assign equipment: %w", err)
	}

	return nil
}
