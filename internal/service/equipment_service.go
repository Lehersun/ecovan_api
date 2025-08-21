package service

import (
	"context"
	"fmt"
	"time"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
)

// equipmentService implements port.EquipmentService
type equipmentService struct {
	equipmentRepo port.EquipmentRepository
}

// NewEquipmentService creates a new equipment service
func NewEquipmentService(equipmentRepo port.EquipmentRepository) port.EquipmentService {
	return &equipmentService{
		equipmentRepo: equipmentRepo,
	}
}

// Create creates a new equipment with validation
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (s *equipmentService) Create(ctx context.Context, req models.CreateEquipmentRequest) (*models.EquipmentResponse, error) {
	// Validate placement (exactly one of client_object_id or warehouse_id)
	if err := req.ValidatePlacement(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if number already exists (if provided)
	if req.Number != nil {
		exists, err := s.equipmentRepo.ExistsByNumber(ctx, *req.Number, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to check equipment number existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("equipment with number '%s' already exists", *req.Number)
		}
	}

	// Create equipment
	equipment := models.FromEquipmentCreateRequest(req)
	equipment.ID = uuid.New()
	equipment.CreatedAt = time.Now()
	equipment.UpdatedAt = time.Now()

	if err := s.equipmentRepo.Create(ctx, &equipment); err != nil {
		return nil, fmt.Errorf("failed to create equipment: %w", err)
	}

	// Return response
	response := equipment.ToResponse()
	return &response, nil
}

// GetByID retrieves equipment by ID
func (s *equipmentService) GetByID(ctx context.Context, id uuid.UUID) (*models.EquipmentResponse, error) {
	equipment, err := s.equipmentRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	response := equipment.ToResponse()
	return &response, nil
}

// List retrieves equipment with pagination and filtering
func (s *equipmentService) List(ctx context.Context, req models.EquipmentListRequest) (*models.EquipmentListResponse, error) {
	// Set defaults if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	response, err := s.equipmentRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list equipment: %w", err)
	}

	return response, nil
}

// Update updates an existing equipment with validation
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (s *equipmentService) Update(ctx context.Context, id uuid.UUID, req models.UpdateEquipmentRequest) (*models.EquipmentResponse, error) {
	// Get existing equipment
	equipment, err := s.equipmentRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Validate update request
	if err := s.validateUpdateRequest(ctx, id, req, equipment); err != nil {
		return nil, err
	}

	// Update equipment
	equipment.UpdateFromRequest(req)

	// Validate placement if specified
	if err := s.validatePlacementUpdate(req); err != nil {
		return nil, err
	}

	equipment.UpdatedAt = time.Now()

	if err := s.equipmentRepo.Update(ctx, equipment); err != nil {
		return nil, fmt.Errorf("failed to update equipment: %w", err)
	}

	// Return response
	response := equipment.ToResponse()
	return &response, nil
}

// validateUpdateRequest validates the update request
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (s *equipmentService) validateUpdateRequest(
	ctx context.Context,
	id uuid.UUID,
	req models.UpdateEquipmentRequest,
	equipment *models.Equipment,
) error {
	// Check if equipment is attached to transport
	isAttached, err := s.equipmentRepo.IsAttachedToTransport(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check transport attachment: %w", err)
	}

	// If placement is being changed and equipment is attached to transport, reject
	if isAttached && s.isPlacementChanging(req, equipment) {
		return fmt.Errorf("cannot change equipment placement while attached to transport")
	}

	// Check if number already exists (if being changed)
	if req.Number != nil && (equipment.Number == nil || *req.Number != *equipment.Number) {
		exists, err := s.equipmentRepo.ExistsByNumber(ctx, *req.Number, &id)
		if err != nil {
			return fmt.Errorf("failed to check equipment number existence: %w", err)
		}
		if exists {
			return fmt.Errorf("equipment with number '%s' already exists", *req.Number)
		}
	}

	return nil
}

// isPlacementChanging checks if the placement is being changed
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (s *equipmentService) isPlacementChanging(req models.UpdateEquipmentRequest, equipment *models.Equipment) bool {
	return req.ClientObjectID != equipment.ClientObjectID ||
		req.WarehouseID != equipment.WarehouseID ||
		req.TransportID != equipment.TransportID
}

// validatePlacementUpdate validates placement update if specified
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (s *equipmentService) validatePlacementUpdate(req models.UpdateEquipmentRequest) error {
	// If no placement was specified in the request, keep the existing placement
	if req.ClientObjectID == nil && req.WarehouseID == nil && req.TransportID == nil {
		return nil // Keep existing placement - no need to change anything
	}

	// Placement was specified, ensure it's valid
	if err := req.ValidatePlacement(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// Delete soft-deletes equipment (only if not attached to transport)
//
//nolint:dupl // Similar pattern across services but with different business logic and checks
func (s *equipmentService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get existing equipment
	equipment, err := s.equipmentRepo.GetByID(ctx, id, false)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}

	// Check if equipment is attached to transport
	isAttached, err := s.equipmentRepo.IsAttachedToTransport(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check transport attachment: %w", err)
	}

	if isAttached {
		return fmt.Errorf("cannot delete equipment while attached to transport")
	}

	// Soft delete equipment
	if err := s.equipmentRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete equipment: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted equipment
func (s *equipmentService) Restore(ctx context.Context, id uuid.UUID) (*models.EquipmentResponse, error) {
	// Get existing equipment (including deleted)
	equipment, err := s.equipmentRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	if equipment.DeletedAt == nil {
		return nil, fmt.Errorf("equipment is not deleted")
	}

	// Check if number conflicts with existing equipment
	if equipment.Number != nil {
		exists, err := s.equipmentRepo.ExistsByNumber(ctx, *equipment.Number, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check equipment number existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("cannot restore equipment: number '%s' conflicts with existing equipment", *equipment.Number)
		}
	}

	// Restore equipment
	if err := s.equipmentRepo.Restore(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to restore equipment: %w", err)
	}

	// Get restored equipment
	restoredEquipment, err := s.equipmentRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored equipment: %w", err)
	}

	if restoredEquipment == nil {
		return nil, fmt.Errorf("failed to get restored equipment")
	}

	// Return response
	response := restoredEquipment.ToResponse()
	return &response, nil
}
