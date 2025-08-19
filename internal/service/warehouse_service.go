package service

import (
	"context"
	"fmt"
	"time"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
)

// warehouseService implements port.WarehouseService
type warehouseService struct {
	warehouseRepo port.WarehouseRepository
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(warehouseRepo port.WarehouseRepository) port.WarehouseService {
	return &warehouseService{
		warehouseRepo: warehouseRepo,
	}
}

// Create creates a new warehouse with validation
func (s *warehouseService) Create(ctx context.Context, req models.CreateWarehouseRequest) (*models.WarehouseResponse, error) {
	// Check if warehouse with same name already exists
	exists, err := s.warehouseRepo.ExistsByName(ctx, req.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check warehouse name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("warehouse with name '%s' already exists", req.Name)
	}

	// Create warehouse model
	warehouse := models.FromWarehouseCreateRequest(req)
	warehouse.ID = uuid.New()
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	// Save to repository
	if err := s.warehouseRepo.Create(ctx, &warehouse); err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}

	// Return response
	response := warehouse.ToResponse()
	return &response, nil
}

// GetByID retrieves a warehouse by ID
func (s *warehouseService) GetByID(ctx context.Context, id uuid.UUID) (*models.WarehouseResponse, error) {
	warehouse, err := s.warehouseRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}
	if warehouse == nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	response := warehouse.ToResponse()
	return &response, nil
}

// List retrieves warehouses with pagination and filtering
func (s *warehouseService) List(ctx context.Context, req models.WarehouseListRequest) (*models.WarehouseListResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	response, err := s.warehouseRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list warehouses: %w", err)
	}

	return response, nil
}

// Update updates an existing warehouse with validation
func (s *warehouseService) Update(ctx context.Context, id uuid.UUID, req models.UpdateWarehouseRequest) (*models.WarehouseResponse, error) {
	// Get existing warehouse
	warehouse, err := s.warehouseRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}
	if warehouse == nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	// Check if new name conflicts with existing warehouse
	exists, err := s.warehouseRepo.ExistsByName(ctx, req.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check warehouse name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("warehouse with name '%s' already exists", req.Name)
	}

	// Update warehouse
	warehouse.UpdateFromRequest(req)
	warehouse.UpdatedAt = time.Now()

	if err := s.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}

	// Return response
	response := warehouse.ToResponse()
	return &response, nil
}

// Delete soft-deletes a warehouse (only if no active equipment)
func (s *warehouseService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if warehouse exists
	warehouse, err := s.warehouseRepo.GetByID(ctx, id, false)
	if err != nil {
		return fmt.Errorf("failed to get warehouse: %w", err)
	}
	if warehouse == nil {
		return fmt.Errorf("warehouse not found")
	}

	// Check if warehouse has active equipment
	hasEquipment, err := s.warehouseRepo.HasActiveEquipment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check warehouse equipment: %w", err)
	}
	if hasEquipment {
		return fmt.Errorf("cannot delete warehouse: equipment is still present")
	}

	// Soft delete warehouse
	if err := s.warehouseRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted warehouse
func (s *warehouseService) Restore(ctx context.Context, id uuid.UUID) (*models.WarehouseResponse, error) {
	// Check if warehouse exists (including deleted)
	warehouse, err := s.warehouseRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}
	if warehouse == nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	// Check if warehouse is already restored
	if warehouse.DeletedAt == nil {
		return nil, fmt.Errorf("warehouse is not deleted")
	}

	// Check if restored name conflicts with existing warehouse
	exists, err := s.warehouseRepo.ExistsByName(ctx, warehouse.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check warehouse name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("cannot restore warehouse: name '%s' conflicts with existing warehouse", warehouse.Name)
	}

	// Restore warehouse
	if err := s.warehouseRepo.Restore(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to restore warehouse: %w", err)
	}

	// Get updated warehouse
	restoredWarehouse, err := s.warehouseRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored warehouse: %w", err)
	}

	// Return response
	response := restoredWarehouse.ToResponse()
	return &response, nil
}
