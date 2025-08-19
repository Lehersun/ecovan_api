package service

import (
	"context"
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

// MockEquipmentRepository is a mock implementation of port.EquipmentRepository
type MockEquipmentRepository struct {
	mock.Mock
}

func (m *MockEquipmentRepository) Create(ctx context.Context, equipment *models.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockEquipmentRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Equipment, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Equipment), args.Error(1)
}

func (m *MockEquipmentRepository) List(ctx context.Context, req models.EquipmentListRequest) (*models.EquipmentListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EquipmentListResponse), args.Error(1)
}

func (m *MockEquipmentRepository) Update(ctx context.Context, equipment *models.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockEquipmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEquipmentRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(1)
}

func (m *MockEquipmentRepository) IsAttachedToTransport(ctx context.Context, equipmentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, equipmentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEquipmentRepository) ExistsByNumber(ctx context.Context, number string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, number, excludeID)
	return args.Bool(0), args.Error(1)
}

func TestNewEquipmentService(t *testing.T) {
	mockRepo := &MockEquipmentRepository{}
	service := NewEquipmentService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.(*equipmentService).equipmentRepo)
}

func TestEquipmentService_Create(t *testing.T) {
	tests := []struct {
		name          string
		req           models.CreateEquipmentRequest
		setupMock     func(*MockEquipmentRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_creation",
			req: models.CreateEquipmentRequest{
				Type:      models.EquipmentTypeBin,
				VolumeL:   100,
				Condition: models.EquipmentConditionGood,
				Number:    stringPtr("BIN001"),
			},
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("ExistsByNumber", mock.Anything, "BIN001", (*uuid.UUID)(nil)).Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Equipment")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "both_locations_set",
			req: models.CreateEquipmentRequest{
				Type:           models.EquipmentTypeContainer,
				VolumeL:        200,
				Condition:      models.EquipmentConditionGood,
				ClientObjectID: uuidPtr(uuid.New()),
				WarehouseID:    uuidPtr(uuid.New()),
			},
			setupMock:     func(repo *MockEquipmentRepository) {},
			expectError:   true,
			expectedError: "validation failed: equipment cannot be placed at both client object and warehouse simultaneously",
		},
		{
			name: "number_already_exists",
			req: models.CreateEquipmentRequest{
				Type:      models.EquipmentTypeBin,
				VolumeL:   100,
				Condition: models.EquipmentConditionGood,
				Number:    stringPtr("BIN001"),
			},
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("ExistsByNumber", mock.Anything, "BIN001", (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectError:   true,
			expectedError: "equipment with number 'BIN001' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEquipmentRepository{}
			tt.setupMock(mockRepo)

			service := NewEquipmentService(mockRepo)
			result, err := service.Create(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Type, result.Type)
				assert.Equal(t, tt.req.VolumeL, result.VolumeL)
				assert.Equal(t, tt.req.Condition, result.Condition)
				assert.Equal(t, tt.req.Number, result.Number)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEquipmentService_Update(t *testing.T) {
	equipmentID := uuid.New()
	existingEquipment := &models.Equipment{
		ID:             equipmentID,
		Type:           string(models.EquipmentTypeBin),
		VolumeL:        100,
		Condition:      string(models.EquipmentConditionGood),
		ClientObjectID: uuidPtr(uuid.New()),
		WarehouseID:    nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	tests := []struct {
		name          string
		req           models.UpdateEquipmentRequest
		setupMock     func(*MockEquipmentRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_update",
			req: models.UpdateEquipmentRequest{
				Type:      models.EquipmentTypeContainer,
				VolumeL:   200,
				Condition: models.EquipmentConditionGood,
			},
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(existingEquipment, nil)
				repo.On("IsAttachedToTransport", mock.Anything, equipmentID).Return(false, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Equipment")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "both_locations_set",
			req: models.UpdateEquipmentRequest{
				Type:           models.EquipmentTypeContainer,
				VolumeL:        200,
				Condition:      models.EquipmentConditionGood,
				ClientObjectID: uuidPtr(uuid.New()),
				WarehouseID:    uuidPtr(uuid.New()),
			},
			setupMock:     func(repo *MockEquipmentRepository) {},
			expectError:   true,
			expectedError: "validation failed: equipment cannot be placed at both client object and warehouse simultaneously",
		},
		{
			name: "equipment_not_found",
			req: models.UpdateEquipmentRequest{
				Type:      models.EquipmentTypeContainer,
				VolumeL:   200,
				Condition: models.EquipmentConditionGood,
			},
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(nil, nil)
			},
			expectError:   true,
			expectedError: "equipment not found",
		},
		{
			name: "change_placement_while_attached_to_transport",
			req: models.UpdateEquipmentRequest{
				Type:        models.EquipmentTypeContainer,
				VolumeL:     200,
				Condition:   models.EquipmentConditionGood,
				WarehouseID: uuidPtr(uuid.New()),
			},
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(existingEquipment, nil)
				repo.On("IsAttachedToTransport", mock.Anything, equipmentID).Return(true, nil)
			},
			expectError:   true,
			expectedError: "cannot change equipment placement while attached to transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEquipmentRepository{}
			tt.setupMock(mockRepo)

			service := NewEquipmentService(mockRepo)
			result, err := service.Update(context.Background(), equipmentID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEquipmentService_Delete(t *testing.T) {
	equipmentID := uuid.New()
	existingEquipment := &models.Equipment{
		ID:        equipmentID,
		Type:      string(models.EquipmentTypeBin),
		VolumeL:   100,
		Condition: string(models.EquipmentConditionGood),
	}

	tests := []struct {
		name          string
		setupMock     func(*MockEquipmentRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_deletion",
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(existingEquipment, nil)
				repo.On("IsAttachedToTransport", mock.Anything, equipmentID).Return(false, nil)
				repo.On("SoftDelete", mock.Anything, equipmentID).Return(nil)
			},
			expectError: false,
		},
		{
			name: "equipment_not_found",
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(nil, nil)
			},
			expectError:   true,
			expectedError: "equipment not found",
		},
		{
			name: "attached_to_transport",
			setupMock: func(repo *MockEquipmentRepository) {
				repo.On("GetByID", mock.Anything, equipmentID, false).Return(existingEquipment, nil)
				repo.On("IsAttachedToTransport", mock.Anything, equipmentID).Return(true, nil)
			},
			expectError:   true,
			expectedError: "cannot delete equipment while attached to transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEquipmentRepository{}
			tt.setupMock(mockRepo)

			service := NewEquipmentService(mockRepo)
			err := service.Delete(context.Background(), equipmentID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
