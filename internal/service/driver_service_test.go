package service

import (
	"context"
	"eco-van-api/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDriverRepository is a mock implementation of DriverRepository
type MockDriverRepository struct {
	mock.Mock
}

func (m *MockDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	args := m.Called(ctx, driver)
	return args.Error(0)
}

func (m *MockDriverRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Driver, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Driver), args.Error(1)
}

func (m *MockDriverRepository) List(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DriverListResponse), args.Error(1)
}

func (m *MockDriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	args := m.Called(ctx, driver)
	return args.Error(0)
}

func (m *MockDriverRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDriverRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDriverRepository) IsAssignedToTransport(ctx context.Context, driverID uuid.UUID) (bool, error) {
	args := m.Called(ctx, driverID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDriverRepository) ExistsByLicenseNo(ctx context.Context, licenseNo string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, licenseNo, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDriverRepository) ListAvailable(ctx context.Context) ([]models.Driver, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Driver), args.Error(1)
}

func TestDriverService_Create(t *testing.T) {
	tests := []struct {
		name          string
		req           models.CreateDriverRequest
		setupMock     func(*MockDriverRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_creation",
			req: models.CreateDriverRequest{
				FullName:     "John Doe",
				Phone:        stringPtr("+1234567890"),
				LicenseNo:    "DL123456789",
				LicenseClass: models.DriverLicenseClassB,
			},
			setupMock: func(repo *MockDriverRepository) {
				repo.On("ExistsByLicenseNo", mock.Anything, "DL123456789", (*uuid.UUID)(nil)).Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Driver")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "license_already_exists",
			req: models.CreateDriverRequest{
				FullName:     "Jane Doe",
				Phone:        stringPtr("+0987654321"),
				LicenseNo:    "DL987654321",
				LicenseClass: models.DriverLicenseClassC,
			},
			setupMock: func(repo *MockDriverRepository) {
				repo.On("ExistsByLicenseNo", mock.Anything, "DL987654321", (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectError:   true,
			expectedError: "driver with license number 'DL987654321' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockDriverRepository{}
			tt.setupMock(mockRepo)

			service := NewDriverService(mockRepo)
			result, err := service.Create(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.FullName, result.FullName)
				assert.Equal(t, tt.req.Phone, result.Phone)
				assert.Equal(t, tt.req.LicenseNo, result.LicenseNo)
				assert.Equal(t, string(tt.req.LicenseClass), result.LicenseClass)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDriverService_Update(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		req           models.UpdateDriverRequest
		setupMock     func(*MockDriverRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_update",
			id:   uuid.New(),
			req: models.UpdateDriverRequest{
				FullName: stringPtr("John Updated"),
				Phone:    stringPtr("+1111111111"),
			},
			setupMock: func(repo *MockDriverRepository) {
				existingDriver := &models.Driver{
					ID:           uuid.New(),
					FullName:     "John Doe",
					Phone:        stringPtr("+1234567890"),
					LicenseNo:    "DL123456789",
					LicenseClass: "B",
				}
				repo.On("GetByID", mock.Anything, mock.Anything, false).Return(existingDriver, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Driver")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "driver_not_found",
			id:   uuid.New(),
			req: models.UpdateDriverRequest{
				FullName: stringPtr("John Updated"),
			},
			setupMock: func(repo *MockDriverRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, false).Return(nil, nil)
			},
			expectError:   true,
			expectedError: "driver not found",
		},
		{
			name: "license_already_exists",
			id:   uuid.New(),
			req: models.UpdateDriverRequest{
				LicenseNo: stringPtr("DL999999999"),
			},
			setupMock: func(repo *MockDriverRepository) {
				existingDriver := &models.Driver{
					ID:           uuid.New(),
					FullName:     "John Doe",
					Phone:        stringPtr("+1234567890"),
					LicenseNo:    "DL123456789",
					LicenseClass: "B",
				}
				repo.On("GetByID", mock.Anything, mock.Anything, false).Return(existingDriver, nil)
				repo.On("ExistsByLicenseNo", mock.Anything, "DL999999999", mock.Anything).Return(true, nil)
			},
			expectError:   true,
			expectedError: "driver with license number 'DL999999999' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockDriverRepository{}
			tt.setupMock(mockRepo)

			service := NewDriverService(mockRepo)
			result, err := service.Update(context.Background(), tt.id, tt.req)

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

func TestDriverService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(*MockDriverRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_delete",
			id:   uuid.New(),
			setupMock: func(repo *MockDriverRepository) {
				repo.On("IsAssignedToTransport", mock.Anything, mock.Anything).Return(false, nil)
				repo.On("SoftDelete", mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "driver_assigned_to_transport",
			id:   uuid.New(),
			setupMock: func(repo *MockDriverRepository) {
				repo.On("IsAssignedToTransport", mock.Anything, mock.Anything).Return(true, nil)
			},
			expectError:   true,
			expectedError: "cannot delete driver while assigned to transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockDriverRepository{}
			tt.setupMock(mockRepo)

			service := NewDriverService(mockRepo)
			err := service.Delete(context.Background(), tt.id)

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

func TestDriverService_Restore(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(*MockDriverRepository)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful_restore",
			id:   uuid.New(),
			setupMock: func(repo *MockDriverRepository) {
				now := time.Now()
				deletedDriver := &models.Driver{
					ID:           uuid.New(),
					FullName:     "John Doe",
					LicenseNo:    "DL123456789",
					LicenseClass: "B",
					DeletedAt:    &now,
				}
				restoredDriver := &models.Driver{
					ID:           deletedDriver.ID,
					FullName:     "John Doe",
					LicenseNo:    "DL123456789",
					LicenseClass: "B",
					DeletedAt:    nil,
				}
				repo.On("GetByID", mock.Anything, mock.Anything, true).Return(deletedDriver, nil)
				repo.On("Restore", mock.Anything, mock.Anything).Return(nil)
				repo.On("GetByID", mock.Anything, mock.Anything, false).Return(restoredDriver, nil)
			},
			expectError: false,
		},
		{
			name: "driver_not_found",
			id:   uuid.New(),
			setupMock: func(repo *MockDriverRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, true).Return(nil, nil)
			},
			expectError:   true,
			expectedError: "driver not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockDriverRepository{}
			tt.setupMock(mockRepo)

			service := NewDriverService(mockRepo)
			result, err := service.Restore(context.Background(), tt.id)

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
