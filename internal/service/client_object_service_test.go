package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"eco-van-api/internal/models"
)

// MockClientObjectRepository is a mock implementation of port.ClientObjectRepository
type MockClientObjectRepository struct {
	mock.Mock
}

func (m *MockClientObjectRepository) Create(ctx context.Context, clientObject *models.ClientObject) error {
	args := m.Called(ctx, clientObject)
	return args.Error(0)
}

func (m *MockClientObjectRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.ClientObject, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientObject), args.Error(1)
}

func (m *MockClientObjectRepository) List(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error) {
	args := m.Called(ctx, clientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientObjectListResponse), args.Error(1)
}

func (m *MockClientObjectRepository) Update(ctx context.Context, clientObject *models.ClientObject) error {
	args := m.Called(ctx, clientObject)
	return args.Error(0)
}

func (m *MockClientObjectRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientObjectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientObjectRepository) ExistsByName(ctx context.Context, clientID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, clientID, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockClientObjectRepository) HasActiveOrders(ctx context.Context, clientObjectID uuid.UUID) (bool, error) {
	args := m.Called(ctx, clientObjectID)
	return args.Bool(0), args.Error(1)
}

func (m *MockClientObjectRepository) HasActiveEquipment(ctx context.Context, clientObjectID uuid.UUID) (bool, error) {
	args := m.Called(ctx, clientObjectID)
	return args.Bool(0), args.Error(1)
}

func (m *MockClientObjectRepository) GetDeleteConflicts(ctx context.Context, clientObjectID uuid.UUID) (*models.DeleteConflicts, error) {
	args := m.Called(ctx, clientObjectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeleteConflicts), args.Error(1)
}

func TestNewClientObjectService(t *testing.T) {
	mockClientObjectRepo := &MockClientObjectRepository{}
	mockClientRepo := &MockClientRepository{}
	service := NewClientObjectService(mockClientObjectRepo, mockClientRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockClientObjectRepo, service.(*clientObjectService).clientObjectRepo)
	assert.Equal(t, mockClientRepo, service.(*clientObjectService).clientRepo)
}

func TestClientObjectService_Create(t *testing.T) {
	clientID := uuid.New()
	tests := []struct {
		name          string
		request       models.CreateClientObjectRequest
		mockSetup     func(*MockClientObjectRepository, *MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful creation",
			request: models.CreateClientObjectRequest{
				Name:    "Test Location",
				Address: "123 Test St",
				GeoLat:  float64Ptr(40.7128),
				GeoLng:  float64Ptr(-74.0060),
				Notes:   stringPtr("Test notes"),
			},
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("ExistsByName", mock.Anything, clientID, "Test Location", (*uuid.UUID)(nil)).Return(false, nil)
				clientObjectRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ClientObject")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "client not found",
			request: models.CreateClientObjectRequest{
				Name:    "Test Location",
				Address: "123 Test St",
			},
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorContains: "client not found",
		},
		{
			name: "name already exists",
			request: models.CreateClientObjectRequest{
				Name:    "Existing Location",
				Address: "123 Test St",
			},
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("ExistsByName", mock.Anything, clientID, "Existing Location", (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectedError: true,
			errorContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientObjectRepo := &MockClientObjectRepository{}
			mockClientRepo := &MockClientRepository{}
			service := NewClientObjectService(mockClientObjectRepo, mockClientRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockClientObjectRepo, mockClientRepo)
			}

			response, err := service.Create(context.Background(), clientID, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Address, response.Address)
			}

			mockClientObjectRepo.AssertExpectations(t)
			mockClientRepo.AssertExpectations(t)
		})
	}
}

func TestClientObjectService_GetByID(t *testing.T) {
	clientID := uuid.New()
	objectID := uuid.New()
	tests := []struct {
		name           string
		clientID       uuid.UUID
		objectID       uuid.UUID
		includeDeleted bool
		mockSetup      func(*MockClientObjectRepository, *MockClientRepository)
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "successful retrieval",
			clientID:       clientID,
			objectID:       objectID,
			includeDeleted: false,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(&models.ClientObject{
					ID:       objectID,
					ClientID: clientID,
					Name:     "Test Location",
					Address:  "123 Test St",
				}, nil)
			},
			expectedError: false,
		},
		{
			name:           "client not found",
			clientID:       clientID,
			objectID:       objectID,
			includeDeleted: false,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorContains: "client not found",
		},
		{
			name:           "client object not found",
			clientID:       clientID,
			objectID:       objectID,
			includeDeleted: false,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorContains: "client object not found",
		},
		{
			name:           "client object belongs to different client",
			clientID:       clientID,
			objectID:       objectID,
			includeDeleted: false,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(&models.ClientObject{
					ID:       objectID,
					ClientID: uuid.New(), // Different client ID
					Name:     "Test Location",
					Address:  "123 Test St",
				}, nil)
			},
			expectedError: true,
			errorContains: "client object not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientObjectRepo := &MockClientObjectRepository{}
			mockClientRepo := &MockClientRepository{}
			service := NewClientObjectService(mockClientObjectRepo, mockClientRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockClientObjectRepo, mockClientRepo)
			}

			response, err := service.GetByID(context.Background(), tt.clientID, tt.objectID, tt.includeDeleted)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.objectID, response.ID)
				assert.Equal(t, tt.clientID, response.ClientID)
			}

			mockClientObjectRepo.AssertExpectations(t)
			mockClientRepo.AssertExpectations(t)
		})
	}
}

func TestClientObjectService_Delete(t *testing.T) {
	clientID := uuid.New()
	objectID := uuid.New()
	tests := []struct {
		name          string
		clientID      uuid.UUID
		objectID      uuid.UUID
		mockSetup     func(*MockClientObjectRepository, *MockClientRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful deletion",
			clientID: clientID,
			objectID: objectID,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(&models.ClientObject{
					ID:       objectID,
					ClientID: clientID,
					Name:     "Test Location",
					Address:  "123 Test St",
				}, nil)
				clientObjectRepo.On("GetDeleteConflicts", mock.Anything, objectID).Return(&models.DeleteConflicts{
					HasActiveOrders:    false,
					HasActiveEquipment: false,
				}, nil)
				clientObjectRepo.On("SoftDelete", mock.Anything, objectID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "has active orders",
			clientID: clientID,
			objectID: objectID,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(&models.ClientObject{
					ID:       objectID,
					ClientID: clientID,
					Name:     "Test Location",
					Address:  "123 Test St",
				}, nil)
				clientObjectRepo.On("GetDeleteConflicts", mock.Anything, objectID).Return(&models.DeleteConflicts{
					HasActiveOrders:    true,
					HasActiveEquipment: false,
					Message:            "Cannot delete client object: has 2 active orders",
				}, nil)
			},
			expectedError: true,
			errorContains: "cannot delete client object",
		},
		{
			name:     "has active equipment",
			clientID: clientID,
			objectID: objectID,
			mockSetup: func(clientObjectRepo *MockClientObjectRepository, clientRepo *MockClientRepository) {
				clientRepo.On("GetByID", mock.Anything, clientID, false).Return(&models.Client{ID: clientID}, nil)
				clientObjectRepo.On("GetByID", mock.Anything, objectID, false).Return(&models.ClientObject{
					ID:       objectID,
					ClientID: clientID,
					Name:     "Test Location",
					Address:  "123 Test St",
				}, nil)
				clientObjectRepo.On("GetDeleteConflicts", mock.Anything, objectID).Return(&models.DeleteConflicts{
					HasActiveOrders:    false,
					HasActiveEquipment: true,
					Message:            "Cannot delete client object: has 1 active equipment",
				}, nil)
			},
			expectedError: true,
			errorContains: "cannot delete client object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientObjectRepo := &MockClientObjectRepository{}
			mockClientRepo := &MockClientRepository{}
			service := NewClientObjectService(mockClientObjectRepo, mockClientRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mockClientObjectRepo, mockClientRepo)
			}

			err := service.Delete(context.Background(), tt.clientID, tt.objectID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockClientObjectRepo.AssertExpectations(t)
			mockClientRepo.AssertExpectations(t)
		})
	}
}

// Helper functions
func float64Ptr(f float64) *float64 {
	return &f
}
