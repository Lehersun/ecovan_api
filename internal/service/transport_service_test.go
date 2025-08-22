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

// MockTransportRepository is a mock implementation of port.TransportRepository
type MockTransportRepository struct {
	mock.Mock
}

func (m *MockTransportRepository) Create(ctx context.Context, transport *models.Transport) error {
	args := m.Called(ctx, transport)
	return args.Error(0)
}

func (m *MockTransportRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Transport, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transport), args.Error(1)
}

func (m *MockTransportRepository) Update(ctx context.Context, transport *models.Transport) error {
	args := m.Called(ctx, transport)
	return args.Error(0)
}

func (m *MockTransportRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransportRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransportRepository) List(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportListResponse), args.Error(1)
}

func (m *MockTransportRepository) ExistsByPlateNo(ctx context.Context, plateNo string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, plateNo, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) HasActiveDriver(ctx context.Context, transportID uuid.UUID) (bool, error) {
	args := m.Called(ctx, transportID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) HasActiveEquipment(ctx context.Context, transportID uuid.UUID) (bool, error) {
	args := m.Called(ctx, transportID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) HasActiveOrders(ctx context.Context, transportID uuid.UUID) (bool, error) {
	args := m.Called(ctx, transportID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) AssignDriver(ctx context.Context, transportID, driverID uuid.UUID) error {
	args := m.Called(ctx, transportID, driverID)
	return args.Error(0)
}

func (m *MockTransportRepository) UnassignDriver(ctx context.Context, transportID uuid.UUID) error {
	args := m.Called(ctx, transportID)
	return args.Error(0)
}

func (m *MockTransportRepository) AssignEquipment(ctx context.Context, transportID, equipmentID uuid.UUID) error {
	args := m.Called(ctx, transportID, equipmentID)
	return args.Error(0)
}

func (m *MockTransportRepository) UnassignEquipment(ctx context.Context, transportID uuid.UUID) error {
	args := m.Called(ctx, transportID)
	return args.Error(0)
}

func (m *MockTransportRepository) GetAvailable(ctx context.Context, req models.TransportListRequest) (
	*models.TransportListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportListResponse), args.Error(1)
}

func (m *MockTransportRepository) IsDriverAssignedToOtherTransport(ctx context.Context, driverID, excludeTransportID uuid.UUID) (
	bool, error) {
	args := m.Called(ctx, driverID, excludeTransportID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) IsEquipmentAssignedToOtherTransport(ctx context.Context, equipmentID, excludeTransportID uuid.UUID) (
	bool, error) {
	args := m.Called(ctx, equipmentID, excludeTransportID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransportRepository) IsEquipmentAvailableForAssignment(ctx context.Context, equipmentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, equipmentID)
	return args.Bool(0), args.Error(1)
}

func TestNewTransportService(t *testing.T) {
	mockTransportRepo := &MockTransportRepository{}
	mockDriverRepo := &MockDriverRepository{}
	mockEquipmentRepo := &MockEquipmentRepository{}

	service := NewTransportService(mockTransportRepo, mockDriverRepo, mockEquipmentRepo)

	assert.NotNil(t, service)
	// Note: We can't test private fields directly, but we can verify the service was created
}

func TestTransportService_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.CreateTransportRequest
		setupMocks     func(*MockTransportRepository, *MockDriverRepository)
		expectedError  string
		expectedResult *models.TransportResponse
	}{
		{
			name: "successful_creation_without_driver",
			request: &models.CreateTransportRequest{
				PlateNo:   "ABC123",
				Brand:     "Toyota",
				Model:     "Camry",
				CapacityL: 1000,
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("ExistsByPlateNo", mock.Anything, "ABC123", (*uuid.UUID)(nil)).Return(false, nil)
				transportRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Transport")).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "successful_creation_with_driver",
			request: &models.CreateTransportRequest{
				PlateNo:   "XYZ789",
				Brand:     "Honda",
				Model:     "Civic",
				CapacityL: 800,
				DriverID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("ExistsByPlateNo", mock.Anything, "XYZ789", (*uuid.UUID)(nil)).Return(false, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Driver{ID: uuid.New()}, nil)
				transportRepo.On("IsDriverAssignedToOtherTransport", mock.Anything, mock.AnythingOfType("uuid.UUID"), uuid.Nil).Return(false, nil)
				transportRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Transport")).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "plate_number_already_exists",
			request: &models.CreateTransportRequest{
				PlateNo:   "EXISTING",
				Brand:     "Ford",
				Model:     "Focus",
				CapacityL: 600,
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("ExistsByPlateNo", mock.Anything, "EXISTING", (*uuid.UUID)(nil)).Return(true, nil)
			},
			expectedError: "transport with plate number EXISTING already exists",
		},
		{
			name: "driver_not_found",
			request: &models.CreateTransportRequest{
				PlateNo:   "NEW123",
				Brand:     "BMW",
				Model:     "X3",
				CapacityL: 1200,
				DriverID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("ExistsByPlateNo", mock.Anything, "NEW123", (*uuid.UUID)(nil)).Return(false, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, nil)
			},
			expectedError: "driver not found",
		},
		{
			name: "driver_already_assigned",
			request: &models.CreateTransportRequest{
				PlateNo:   "NEW456",
				Brand:     "Audi",
				Model:     "A4",
				CapacityL: 900,
				DriverID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("ExistsByPlateNo", mock.Anything, "NEW456", (*uuid.UUID)(nil)).Return(false, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Driver{ID: uuid.New()}, nil)
				transportRepo.On("IsDriverAssignedToOtherTransport", mock.Anything, mock.AnythingOfType("uuid.UUID"), uuid.Nil).Return(true, nil)
			},
			expectedError: "driver is already assigned to another transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransportRepo := &MockTransportRepository{}
			mockDriverRepo := &MockDriverRepository{}
			mockEquipmentRepo := &MockEquipmentRepository{}

			service := NewTransportService(mockTransportRepo, mockDriverRepo, mockEquipmentRepo)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTransportRepo, mockDriverRepo)
			}

			result, err := service.Create(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.PlateNo, result.PlateNo)
				assert.Equal(t, tt.request.Brand, result.Brand)
				assert.Equal(t, tt.request.Model, result.Model)
				assert.Equal(t, tt.request.CapacityL, result.CapacityL)
				assert.Equal(t, "IN_WORK", result.Status)

				if tt.request.DriverID != nil {
					assert.Equal(t, tt.request.DriverID, result.CurrentDriverID)
				} else {
					assert.Nil(t, result.CurrentDriverID)
				}
			}

			mockTransportRepo.AssertExpectations(t)
			mockDriverRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create UpdateTransportRequest with proper explicit flags
func createUpdateRequestWithDriver(driverID *uuid.UUID) models.UpdateTransportRequest {
	req := models.UpdateTransportRequest{}
	if driverID != nil {
		req.DriverID = driverID
		req.DriverIDExplicitlySet = true
		req.DriverIDExplicitlyNull = false
	}
	return req
}

// Helper function to create UpdateTransportRequest for unassigning driver
func createUpdateRequestUnassignDriver() models.UpdateTransportRequest {
	req := models.UpdateTransportRequest{}
	req.DriverIDExplicitlySet = true
	req.DriverIDExplicitlyNull = true
	return req
}

func TestTransportService_Update(t *testing.T) {
	tests := []struct {
		name              string
		transportID       uuid.UUID
		request           models.UpdateTransportRequest
		existingTransport *models.Transport
		setupMocks        func(*MockTransportRepository, *MockDriverRepository)
		expectedError     string
		expectedResult    *models.TransportResponse
	}{
		{
			name:        "successful_update_without_driver_change",
			transportID: uuid.New(),
			request: models.UpdateTransportRequest{
				Brand: stringPtr("Updated Brand"),
				Model: stringPtr("Updated Model"),
			},
			existingTransport: &models.Transport{
				ID:              uuid.New(),
				PlateNo:         "ABC123",
				Brand:           "Old Brand",
				Model:           "Old Model",
				CapacityL:       1000,
				Status:          "IN_WORK",
				CurrentDriverID: nil,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Transport{
					ID:              uuid.New(),
					PlateNo:         "ABC123",
					Brand:           "Old Brand",
					Model:           "Old Model",
					CapacityL:       1000,
					Status:          "IN_WORK",
					CurrentDriverID: nil,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}, nil)
				transportRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Transport")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "successful_update_assigning_driver",
			transportID: uuid.New(),
			request: func() models.UpdateTransportRequest {
				req := createUpdateRequestWithDriver(func() *uuid.UUID { id := uuid.New(); return &id }())
				req.Brand = stringPtr("Updated Brand")
				return req
			}(),
			existingTransport: &models.Transport{
				ID:              uuid.New(),
				PlateNo:         "ABC123",
				Brand:           "Old Brand",
				Model:           "Old Model",
				CapacityL:       1000,
				Status:          "IN_WORK",
				CurrentDriverID: nil,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Transport{
					ID:              uuid.New(),
					PlateNo:         "ABC123",
					Brand:           "Old Brand",
					Model:           "Old Model",
					CapacityL:       1000,
					Status:          "IN_WORK",
					CurrentDriverID: nil,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Driver{ID: uuid.New()}, nil)
				transportRepo.On("IsDriverAssignedToOtherTransport", mock.Anything, mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType("uuid.UUID")).Return(false, nil)
				transportRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Transport")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "successful_update_unassigning_driver",
			transportID: uuid.New(),
			request: func() models.UpdateTransportRequest {
				req := createUpdateRequestUnassignDriver()
				req.Brand = stringPtr("Updated Brand")
				return req
			}(),
			existingTransport: &models.Transport{
				ID:              uuid.New(),
				PlateNo:         "ABC123",
				Brand:           "Old Brand",
				Model:           "Old Model",
				CapacityL:       1000,
				Status:          "IN_WORK",
				CurrentDriverID: &uuid.UUID{},
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Transport{
					ID:              uuid.New(),
					PlateNo:         "ABC123",
					Brand:           "Old Brand",
					Model:           "Old Model",
					CapacityL:       1000,
					Status:          "IN_WORK",
					CurrentDriverID: &uuid.UUID{},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}, nil)
				transportRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Transport")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "driver_not_found",
			transportID: uuid.New(),
			request: func() models.UpdateTransportRequest {
				req := createUpdateRequestWithDriver(func() *uuid.UUID { id := uuid.New(); return &id }())
				req.Brand = stringPtr("Updated Brand")
				return req
			}(),
			existingTransport: &models.Transport{
				ID:              uuid.New(),
				PlateNo:         "ABC123",
				Brand:           "Old Brand",
				Model:           "Old Model",
				CapacityL:       1000,
				Status:          "IN_WORK",
				CurrentDriverID: nil,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Transport{
					ID:              uuid.New(),
					PlateNo:         "ABC123",
					Brand:           "Old Brand",
					Model:           "Old Model",
					CapacityL:       1000,
					Status:          "IN_WORK",
					CurrentDriverID: nil,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(nil, nil)
			},
			expectedError: "driver not found",
		},
		{
			name:        "driver_already_assigned_to_other_transport",
			transportID: uuid.New(),
			request: func() models.UpdateTransportRequest {
				req := createUpdateRequestWithDriver(func() *uuid.UUID { id := uuid.New(); return &id }())
				req.Brand = stringPtr("Updated Brand")
				return req
			}(),
			existingTransport: &models.Transport{
				ID:              uuid.New(),
				PlateNo:         "ABC123",
				Brand:           "Old Brand",
				Model:           "Old Model",
				CapacityL:       1000,
				Status:          "IN_WORK",
				CurrentDriverID: nil,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			setupMocks: func(transportRepo *MockTransportRepository, driverRepo *MockDriverRepository) {
				transportRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Transport{
					ID:              uuid.New(),
					PlateNo:         "ABC123",
					Brand:           "Old Brand",
					Model:           "Old Model",
					CapacityL:       1000,
					Status:          "IN_WORK",
					CurrentDriverID: nil,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}, nil)
				driverRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), false).Return(&models.Driver{ID: uuid.New()}, nil)
				transportRepo.On("IsDriverAssignedToOtherTransport", mock.Anything, mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType("uuid.UUID")).Return(true, nil)
			},
			expectedError: "driver is already assigned to another transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransportRepo := &MockTransportRepository{}
			mockDriverRepo := &MockDriverRepository{}
			mockEquipmentRepo := &MockEquipmentRepository{}

			service := NewTransportService(mockTransportRepo, mockDriverRepo, mockEquipmentRepo)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTransportRepo, mockDriverRepo)
			}

			result, err := service.Update(context.Background(), tt.transportID, tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the updated fields
				if tt.request.Brand != nil {
					assert.Equal(t, *tt.request.Brand, result.Brand)
				}
				if tt.request.Model != nil {
					assert.Equal(t, *tt.request.Model, result.Model)
				}
				if tt.request.DriverID != nil {
					assert.Equal(t, tt.request.DriverID, result.CurrentDriverID)
				}
			}

			mockTransportRepo.AssertExpectations(t)
			mockDriverRepo.AssertExpectations(t)
		})
	}
}

func TestTransportService_Create_Integration(t *testing.T) {
	t.Run("create_transport_with_driver_assignment", func(t *testing.T) {
		mockTransportRepo := &MockTransportRepository{}
		mockDriverRepo := &MockDriverRepository{}
		mockEquipmentRepo := &MockEquipmentRepository{}

		service := NewTransportService(mockTransportRepo, mockDriverRepo, mockEquipmentRepo)

		driverID := uuid.New()
		request := &models.CreateTransportRequest{
			PlateNo:   "INT123",
			Brand:     "Tesla",
			Model:     "Model 3",
			CapacityL: 1500,
			DriverID:  &driverID,
		}

		// Setup mocks for successful creation with driver assignment
		mockTransportRepo.On("ExistsByPlateNo", mock.Anything, "INT123", (*uuid.UUID)(nil)).Return(false, nil)
		mockDriverRepo.On("GetByID", mock.Anything, driverID, false).Return(&models.Driver{ID: driverID}, nil)
		mockTransportRepo.On("IsDriverAssignedToOtherTransport", mock.Anything, driverID, uuid.Nil).Return(false, nil)
		mockTransportRepo.On("Create", mock.Anything, mock.MatchedBy(func(transport *models.Transport) bool {
			return transport.PlateNo == "INT123" &&
				transport.Brand == "Tesla" &&
				transport.Model == "Model 3" &&
				transport.CapacityL == 1500 &&
				transport.Status == "IN_WORK" &&
				transport.CurrentDriverID != nil &&
				*transport.CurrentDriverID == driverID
		})).Return(nil)

		result, err := service.Create(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "INT123", result.PlateNo)
		assert.Equal(t, "Tesla", result.Brand)
		assert.Equal(t, "Model 3", result.Model)
		assert.Equal(t, 1500, result.CapacityL)
		assert.Equal(t, "IN_WORK", result.Status)
		assert.Equal(t, &driverID, result.CurrentDriverID)

		mockTransportRepo.AssertExpectations(t)
		mockDriverRepo.AssertExpectations(t)
	})

	t.Run("create_transport_without_driver_assignment", func(t *testing.T) {
		mockTransportRepo := &MockTransportRepository{}
		mockDriverRepo := &MockDriverRepository{}
		mockEquipmentRepo := &MockEquipmentRepository{}

		service := NewTransportService(mockTransportRepo, mockDriverRepo, mockEquipmentRepo)

		request := &models.CreateTransportRequest{
			PlateNo:   "NO_DRIVER",
			Brand:     "Volkswagen",
			Model:     "Golf",
			CapacityL: 700,
		}

		// Setup mocks for successful creation without driver
		mockTransportRepo.On("ExistsByPlateNo", mock.Anything, "NO_DRIVER", (*uuid.UUID)(nil)).Return(false, nil)
		mockTransportRepo.On("Create", mock.Anything, mock.MatchedBy(func(transport *models.Transport) bool {
			return transport.PlateNo == "NO_DRIVER" &&
				transport.Brand == "Volkswagen" &&
				transport.Model == "Golf" &&
				transport.CapacityL == 700 &&
				transport.Status == "IN_WORK" &&
				transport.CurrentDriverID == nil
		})).Return(nil)

		result, err := service.Create(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "NO_DRIVER", result.PlateNo)
		assert.Equal(t, "Volkswagen", result.Brand)
		assert.Equal(t, "Golf", result.Model)
		assert.Equal(t, 700, result.CapacityL)
		assert.Equal(t, "IN_WORK", result.Status)
		assert.Nil(t, result.CurrentDriverID)

		mockTransportRepo.AssertExpectations(t)
	})
}
