package service

import (
	"context"
	"eco-van-api/internal/models"
	"eco-van-api/internal/port"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type driverService struct {
	driverRepo port.DriverRepository
}

// NewDriverService creates a new driver service
func NewDriverService(driverRepo port.DriverRepository) port.DriverService {
	return &driverService{driverRepo: driverRepo}
}

// Create creates a new driver with validation
func (s *driverService) Create(ctx context.Context, req models.CreateDriverRequest) (*models.DriverResponse, error) {
	// Check if license number already exists
	exists, err := s.driverRepo.ExistsByLicenseNo(ctx, req.LicenseNo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check driver license existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("driver with license number '%s' already exists", req.LicenseNo)
	}

	// Create driver
	driver := models.FromDriverCreateRequest(req)
	driver.ID = uuid.New()

	if err := s.driverRepo.Create(ctx, driver); err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Return response
	response := driver.ToResponse()
	return &response, nil
}

// GetByID retrieves driver by ID
func (s *driverService) GetByID(ctx context.Context, id uuid.UUID) (*models.DriverResponse, error) {
	driver, err := s.driverRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	if driver == nil {
		return nil, fmt.Errorf("driver not found")
	}

	// Return response
	response := driver.ToResponse()
	return &response, nil
}

// List retrieves drivers with pagination and filtering
func (s *driverService) List(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error) {
	response, err := s.driverRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list drivers: %w", err)
	}

	return response, nil
}

// Update updates an existing driver with validation
func (s *driverService) Update(ctx context.Context, id uuid.UUID, req models.UpdateDriverRequest) (*models.DriverResponse, error) {
	// Get existing driver
	driver, err := s.driverRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	if driver == nil {
		return nil, fmt.Errorf("driver not found")
	}

	// Check if license number already exists (if being changed)
	if req.LicenseNo != nil && *req.LicenseNo != driver.LicenseNo {
		exists, err := s.driverRepo.ExistsByLicenseNo(ctx, *req.LicenseNo, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check driver license existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("driver with license number '%s' already exists", *req.LicenseNo)
		}
	}

	// Update driver
	driver.UpdateFromRequest(req)
	driver.UpdatedAt = time.Now()

	if err := s.driverRepo.Update(ctx, driver); err != nil {
		return nil, fmt.Errorf("failed to update driver: %w", err)
	}

	// Return response
	response := driver.ToResponse()
	return &response, nil
}

// Delete soft-deletes driver (only if not assigned to transport)
func (s *driverService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if driver is assigned to transport
	isAssigned, err := s.driverRepo.IsAssignedToTransport(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check driver transport assignment: %w", err)
	}

	if isAssigned {
		return fmt.Errorf("cannot delete driver while assigned to transport")
	}

	// Soft delete driver
	if err := s.driverRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete driver: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted driver
func (s *driverService) Restore(ctx context.Context, id uuid.UUID) (*models.DriverResponse, error) {
	// Get driver (including soft-deleted)
	driver, err := s.driverRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	if driver == nil {
		return nil, fmt.Errorf("driver not found")
	}

	if driver.DeletedAt == nil {
		return nil, fmt.Errorf("driver is not soft-deleted")
	}

	// Restore driver
	if err := s.driverRepo.Restore(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to restore driver: %w", err)
	}

	// Get updated driver
	driver, err = s.driverRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored driver: %w", err)
	}

	// Return response
	response := driver.ToResponse()
	return &response, nil
}

// ListAvailable retrieves available drivers (not assigned to any transport)
func (s *driverService) ListAvailable(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error) {
	response, err := s.driverRepo.ListAvailable(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list available drivers: %w", err)
	}

	return response, nil
}
