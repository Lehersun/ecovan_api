package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"eco-van-api/internal/models"
)

// MockClientRepository is a mock implementation of port.ClientRepository
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Client, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) List(ctx context.Context, req models.ClientListRequest) (*models.ClientListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientListResponse), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func TestNewClientService(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.(*clientService).clientRepo)
}

func TestClientService_Create(t *testing.T) {
	tests := []struct {
		name          string
		request       models.CreateClientRequest
		mockSetup     func(*MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful creation",
			request: models.CreateClientRequest{
				Name:  "Test Company",
				TaxID: stringPtr("123456789"),
				Email: stringPtr("test@company.com"),
				Phone: stringPtr("+1234567890"),
				Notes: stringPtr("Test notes"),
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("ExistsByName", mock.Anything, "Test Company", (*uuid.UUID)(nil)).Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Client")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "name already exists",
			request: models.CreateClientRequest{
				Name: "Existing Company",
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("ExistsByName", mock.Anything, "Existing Company", (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectedError: true,
			errorContains: "already exists",
		},
		{
			name: "repository error on name check",
			request: models.CreateClientRequest{
				Name: "Test Company",
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("ExistsByName", mock.Anything, "Test Company", (*uuid.UUID)(nil)).Return(false, assert.AnError)
			},
			expectedError: true,
			errorContains: "failed to check client name existence",
		},
		{
			name: "repository error on create",
			request: models.CreateClientRequest{
				Name: "Test Company",
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("ExistsByName", mock.Anything, "Test Company", (*uuid.UUID)(nil)).Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Client")).Return(assert.AnError)
			},
			expectedError: true,
			errorContains: "failed to create client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			result, err := service.Create(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.TaxID, result.TaxID)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.Phone, result.Phone)
				assert.Equal(t, tt.request.Notes, result.Notes)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		clientID      uuid.UUID
		mockSetup     func(*MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful retrieval",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				client := &models.Client{
					ID:    uuid.New(),
					Name:  "Test Company",
					TaxID: stringPtr("123456789"),
					Email: stringPtr("test@company.com"),
					Phone: stringPtr("+1234567890"),
					Notes: stringPtr("Test notes"),
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(client, nil)
			},
			expectedError: false,
		},
		{
			name:     "client not found",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:     "repository error",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorContains: "failed to get client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			result, err := service.GetByID(context.Background(), tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientService_List(t *testing.T) {
	tests := []struct {
		name          string
		request       models.ClientListRequest
		mockSetup     func(*MockClientRepository)
		expectedError bool
	}{
		{
			name: "successful list with defaults",
			request: models.ClientListRequest{
				Page:     0,
				PageSize: 0,
			},
			mockSetup: func(repo *MockClientRepository) {
				response := &models.ClientListResponse{
					Items:    []models.Client{},
					Page:     1,
					PageSize: 20,
					Total:    0,
				}
				repo.On("List", mock.Anything, mock.AnythingOfType("models.ClientListRequest")).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name: "successful list with custom values",
			request: models.ClientListRequest{
				Page:           5,
				PageSize:       50,
				Query:          "test",
				IncludeDeleted: true,
			},
			mockSetup: func(repo *MockClientRepository) {
				response := &models.ClientListResponse{
					Items:    []models.Client{},
					Page:     5,
					PageSize: 50,
					Total:    0,
				}
				repo.On("List", mock.Anything, mock.AnythingOfType("models.ClientListRequest")).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name: "page size too large",
			request: models.ClientListRequest{
				Page:     1,
				PageSize: 150,
			},
			mockSetup: func(repo *MockClientRepository) {
				response := &models.ClientListResponse{
					Items:    []models.Client{},
					Page:     1,
					PageSize: 100, // Should be capped at 100
					Total:    0,
				}
				repo.On("List", mock.Anything, mock.AnythingOfType("models.ClientListRequest")).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name: "repository error",
			request: models.ClientListRequest{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("List", mock.Anything, mock.AnythingOfType("models.ClientListRequest")).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			result, err := service.List(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientService_Update(t *testing.T) {
	tests := []struct {
		name          string
		clientID      uuid.UUID
		request       models.UpdateClientRequest
		mockSetup     func(*MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful update",
			clientID: uuid.New(),
			request: models.UpdateClientRequest{
				Name:  "Updated Company",
				TaxID: stringPtr("987654321"),
				Email: stringPtr("updated@company.com"),
				Phone: stringPtr("+9876543210"),
				Notes: stringPtr("Updated notes"),
			},
			mockSetup: func(repo *MockClientRepository) {
				existingClient := &models.Client{
					ID:    uuid.New(),
					Name:  "Old Company",
					TaxID: stringPtr("123456789"),
					Email: stringPtr("old@company.com"),
					Phone: stringPtr("+1234567890"),
					Notes: stringPtr("Old notes"),
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(existingClient, nil)
				repo.On("ExistsByName", mock.Anything, "Updated Company", mock.AnythingOfType("*uuid.UUID")).Return(false, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Client")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "client not found",
			clientID: uuid.New(),
			request: models.UpdateClientRequest{
				Name: "Updated Company",
			},
			mockSetup: func(repo *MockClientRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:     "name already exists",
			clientID: uuid.New(),
			request: models.UpdateClientRequest{
				Name: "Existing Company",
			},
			mockSetup: func(repo *MockClientRepository) {
				existingClient := &models.Client{
					ID:   uuid.New(),
					Name: "Old Company",
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(existingClient, nil)
				repo.On("ExistsByName", mock.Anything, "Existing Company", mock.AnythingOfType("*uuid.UUID")).Return(true, nil)
			},
			expectedError: true,
			errorContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			result, err := service.Update(context.Background(), tt.clientID, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		clientID      uuid.UUID
		mockSetup     func(*MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful delete",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				existingClient := &models.Client{
					ID:   uuid.New(),
					Name: "Test Company",
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(existingClient, nil)
				repo.On("SoftDelete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "client not found",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			err := service.Delete(context.Background(), tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientService_Restore(t *testing.T) {
	tests := []struct {
		name          string
		clientID      uuid.UUID
		mockSetup     func(*MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful restore",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				deletedClient := &models.Client{
					ID:        uuid.New(),
					Name:      "Deleted Company",
					DeletedAt: timePtr(time.Now()),
				}
				restoredClient := &models.Client{
					ID:        uuid.New(),
					Name:      "Deleted Company",
					DeletedAt: nil,
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(deletedClient, nil)
				repo.On("ExistsByName", mock.Anything, "Deleted Company", mock.AnythingOfType("*uuid.UUID")).Return(false, nil)
				repo.On("Restore", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(restoredClient, nil)
			},
			expectedError: false,
		},
		{
			name:     "client not found",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:     "client not deleted",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				activeClient := &models.Client{
					ID:        uuid.New(),
					Name:      "Active Company",
					DeletedAt: nil,
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(activeClient, nil)
			},
			expectedError: true,
			errorContains: "not deleted",
		},
		{
			name:     "name already taken",
			clientID: uuid.New(),
			mockSetup: func(repo *MockClientRepository) {
				deletedClient := &models.Client{
					ID:        uuid.New(),
					Name:      "Deleted Company",
					DeletedAt: timePtr(time.Now()),
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(deletedClient, nil)
				repo.On("ExistsByName", mock.Anything, "Deleted Company", mock.AnythingOfType("*uuid.UUID")).Return(true, nil)
			},
			expectedError: true,
			errorContains: "already taken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			service := NewClientService(mockRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			result, err := service.Restore(context.Background(), tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
