package service

import (
	"context"
	"fmt"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type clientObjectService struct {
	clientObjectRepo port.ClientObjectRepository
	clientRepo       port.ClientRepository
	validate         *validator.Validate
}

// NewClientObjectService creates a new client object service
func NewClientObjectService(clientObjectRepo port.ClientObjectRepository, clientRepo port.ClientRepository) port.ClientObjectService {
	return &clientObjectService{
		clientObjectRepo: clientObjectRepo,
		clientRepo:       clientRepo,
		validate:         validator.New(),
	}
}

// Create creates a new client object for a client
func (s *clientObjectService) Create(ctx context.Context, clientID uuid.UUID, req models.CreateClientObjectRequest) (*models.ClientObjectResponse, error) {
	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Check if name already exists for this client
	exists, err := s.clientObjectRepo.ExistsByName(ctx, clientID, req.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("client object with name '%s' already exists for this client", req.Name)
	}

	// Create client object
	clientObject := models.FromCreateClientObjectRequest(clientID, req)
	if err := s.clientObjectRepo.Create(ctx, clientObject); err != nil {
		return nil, fmt.Errorf("failed to create client object: %w", err)
	}

	response := clientObject.ToResponse()
	return &response, nil
}

// GetByID retrieves a client object by ID
func (s *clientObjectService) GetByID(ctx context.Context, clientID, id uuid.UUID, includeDeleted bool) (*models.ClientObjectResponse, error) {
	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Get client object
	clientObject, err := s.clientObjectRepo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("client object not found: %w", err)
	}

	// Verify the client object belongs to the specified client
	if clientObject.ClientID != clientID {
		return nil, fmt.Errorf("client object not found")
	}

	response := clientObject.ToResponse()
	return &response, nil
}

// List retrieves client objects for a specific client with pagination
func (s *clientObjectService) List(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error) {
	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Get client objects
	response, err := s.clientObjectRepo.ListByClient(ctx, clientID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list client objects: %w", err)
	}

	return response, nil
}

// Update updates an existing client object
func (s *clientObjectService) Update(ctx context.Context, clientID, id uuid.UUID, req models.UpdateClientObjectRequest) (*models.ClientObjectResponse, error) {
	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Get existing client object
	clientObject, err := s.clientObjectRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("client object not found: %w", err)
	}

	// Verify the client object belongs to the specified client
	if clientObject.ClientID != clientID {
		return nil, fmt.Errorf("client object not found")
	}

	// Check if name already exists for this client (excluding current object)
	exists, err := s.clientObjectRepo.ExistsByName(ctx, clientID, req.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("client object with name '%s' already exists for this client", req.Name)
	}

	// Update client object
	clientObject.UpdateFromRequest(req)
	if err := s.clientObjectRepo.Update(ctx, clientObject); err != nil {
		return nil, fmt.Errorf("failed to update client object: %w", err)
	}

	response := clientObject.ToResponse()
	return &response, nil
}

// Delete soft deletes a client object (guarded)
func (s *clientObjectService) Delete(ctx context.Context, clientID, id uuid.UUID) error {
	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Get existing client object
	clientObject, err := s.clientObjectRepo.GetByID(ctx, id, false)
	if err != nil {
		return fmt.Errorf("client object not found: %w", err)
	}

	// Verify the client object belongs to the specified client
	if clientObject.ClientID != clientID {
		return fmt.Errorf("client object not found")
	}

	// Check for conflicts before deletion
	conflicts, err := s.clientObjectRepo.GetDeleteConflicts(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check delete conflicts: %w", err)
	}

	if conflicts.HasActiveOrders || conflicts.HasActiveEquipment {
		return fmt.Errorf("cannot delete client object: %s", conflicts.Message)
	}

	// Soft delete client object
	if err := s.clientObjectRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete client object: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted client object
func (s *clientObjectService) Restore(ctx context.Context, clientID, id uuid.UUID) (*models.ClientObjectResponse, error) {
	// Verify client exists
	_, err := s.clientRepo.GetByID(ctx, clientID, false)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Get existing client object (including deleted)
	clientObject, err := s.clientObjectRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("client object not found: %w", err)
	}

	// Verify the client object belongs to the specified client
	if clientObject.ClientID != clientID {
		return nil, fmt.Errorf("client object not found")
	}

	// Check if name already exists for this client (among non-deleted objects)
	exists, err := s.clientObjectRepo.ExistsByName(ctx, clientID, clientObject.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("client object with name '%s' already exists for this client", clientObject.Name)
	}

	// Restore client object
	if err := s.clientObjectRepo.Restore(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to restore client object: %w", err)
	}

	// Get updated client object
	restoredObject, err := s.clientObjectRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored client object: %w", err)
	}

	response := restoredObject.ToResponse()
	return &response, nil
}
