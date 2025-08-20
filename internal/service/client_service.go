package service

import (
	"context"
	"fmt"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
)

// clientService implements port.ClientService
type clientService struct {
	clientRepo port.ClientRepository
}

// NewClientService creates a new client service
func NewClientService(clientRepo port.ClientRepository) port.ClientService {
	return &clientService{
		clientRepo: clientRepo,
	}
}

// Create creates a new client with validation
//
//nolint:dupl // Similar pattern across services but with different business logic
func (s *clientService) Create(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	// Check if client with same name already exists
	exists, err := s.clientRepo.ExistsByName(ctx, req.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check client name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("client with name '%s' already exists", req.Name)
	}

	// Create client from request
	client := models.FromCreateRequest(req)

	// Save to repository
	err = s.clientRepo.Create(ctx, &client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Return response
	response := client.ToResponse()
	return &response, nil
}

// GetByID retrieves a client by ID
func (s *clientService) GetByID(ctx context.Context, id uuid.UUID) (*models.ClientResponse, error) {
	client, err := s.clientRepo.GetByID(ctx, id, false) // Don't include soft-deleted by default
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	response := client.ToResponse()
	return &response, nil
}

// List retrieves clients with pagination and filtering
func (s *clientService) List(ctx context.Context, req models.ClientListRequest) (*models.ClientListResponse, error) {
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

	response, err := s.clientRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	return response, nil
}

// Update updates an existing client with validation
func (s *clientService) Update(ctx context.Context, id uuid.UUID, req models.UpdateClientRequest) (*models.ClientResponse, error) {
	// Get existing client
	client, err := s.clientRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	// Check if new name conflicts with existing client (excluding current one)
	exists, err := s.clientRepo.ExistsByName(ctx, req.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("client with name '%s' already exists", req.Name)
	}

	// Update client from request
	client.UpdateFromRequest(req)

	// Save to repository
	err = s.clientRepo.Update(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	// Return updated response
	response := client.ToResponse()
	return &response, nil
}

// Delete soft-deletes a client
func (s *clientService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if client exists and is not already deleted
	client, err := s.clientRepo.GetByID(ctx, id, false)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	if client == nil {
		return fmt.Errorf("client not found")
	}

	// Soft delete the client
	err = s.clientRepo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted client
func (s *clientService) Restore(ctx context.Context, id uuid.UUID) (*models.ClientResponse, error) {
	// Check if client exists and is deleted
	client, err := s.clientRepo.GetByID(ctx, id, true) // Include soft-deleted
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	if client.DeletedAt == nil {
		return nil, fmt.Errorf("client is not deleted")
	}

	// Check if name conflicts with existing active client
	exists, err := s.clientRepo.ExistsByName(ctx, client.Name, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("cannot restore client: name '%s' is already taken by another client", client.Name)
	}

	// Restore the client
	err = s.clientRepo.Restore(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to restore client: %w", err)
	}

	// Get the restored client
	restoredClient, err := s.clientRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored client: %w", err)
	}

	response := restoredClient.ToResponse()
	return &response, nil
}
