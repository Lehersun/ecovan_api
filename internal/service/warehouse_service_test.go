package service

import (
	"context"
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWarehouseRepository is a mock implementation of port.WarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Warehouse, error) {
	args := m.Called(ctx, id, includeDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) List(ctx context.Context, req models.WarehouseListRequest) (*models.WarehouseListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WarehouseListResponse), args.Error(1)
}

func (m *MockWarehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWarehouseRepository) ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWarehouseRepository) HasActiveEquipment(ctx context.Context, warehouseID uuid.UUID) (bool, error) {
	args := m.Called(ctx, warehouseID)
	return args.Bool(0), args.Error(1)
}

func TestWarehouseService_Create(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		req := models.CreateWarehouseRequest{
			Name:    "Test Warehouse",
			Address: stringPtr("123 Test St"),
			Notes:   stringPtr("Test notes"),
		}

		repo.On("ExistsByName", ctx, "Test Warehouse", (*uuid.UUID)(nil)).Return(false, nil)
		repo.On("Create", ctx, mock.AnythingOfType("*models.Warehouse")).Return(nil)

		response, err := service.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Test Warehouse", response.Name)
		assert.Equal(t, "123 Test St", *response.Address)
		assert.Equal(t, "Test notes", *response.Notes)

		repo.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		req := models.CreateWarehouseRequest{
			Name: "Existing Warehouse",
		}

		repo.On("ExistsByName", ctx, "Existing Warehouse", (*uuid.UUID)(nil)).Return(true, nil)

		response, err := service.Create(ctx, req)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "already exists")

		repo.AssertExpectations(t)
	})
}

func TestWarehouseService_GetByID(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		warehouseID := uuid.New()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Test Warehouse",
			Address:   stringPtr("123 Test St"),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(warehouse, nil)

		response, err := service.GetByID(ctx, warehouseID)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, warehouseID, response.ID)
		assert.Equal(t, "Test Warehouse", response.Name)

		repo.AssertExpectations(t)
	})

	t.Run("warehouse not found", func(t *testing.T) {
		warehouseID := uuid.New()

		repo.On("GetByID", ctx, warehouseID, false).Return(nil, nil)

		response, err := service.GetByID(ctx, warehouseID)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "not found")

		repo.AssertExpectations(t)
	})
}

func TestWarehouseService_List(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		req := models.WarehouseListRequest{
			Page:           1,
			PageSize:       10,
			IncludeDeleted: false,
		}

		response := &models.WarehouseListResponse{
			Items:    []models.Warehouse{},
			Page:     1,
			PageSize: 10,
			Total:    0,
		}

		repo.On("List", ctx, req).Return(response, nil)

		result, err := service.List(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, response, result)

		repo.AssertExpectations(t)
	})

	t.Run("sets defaults", func(t *testing.T) {
		req := models.WarehouseListRequest{
			Page:     0,
			PageSize: 0,
		}

		expectedReq := models.WarehouseListRequest{
			Page:           1,
			PageSize:       20,
			IncludeDeleted: false,
		}

		response := &models.WarehouseListResponse{
			Items:    []models.Warehouse{},
			Page:     1,
			PageSize: 20,
			Total:    0,
		}

		repo.On("List", ctx, expectedReq).Return(response, nil)

		result, err := service.List(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, response, result)

		repo.AssertExpectations(t)
	})
}

func TestWarehouseService_Update(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		warehouseID := uuid.New()
		req := models.UpdateWarehouseRequest{
			Name:    "Updated Warehouse",
			Address: stringPtr("Updated Address"),
		}

		existingWarehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Original Name",
			Address:   stringPtr("Original Address"),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(existingWarehouse, nil)
		repo.On("ExistsByName", ctx, "Updated Warehouse", &warehouseID).Return(false, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*models.Warehouse")).Return(nil)

		response, err := service.Update(ctx, warehouseID, req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Updated Warehouse", response.Name)

		repo.AssertExpectations(t)
	})

	t.Run("warehouse not found", func(t *testing.T) {
		warehouseID := uuid.New()
		req := models.UpdateWarehouseRequest{
			Name: "Updated Warehouse",
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(nil, nil)

		response, err := service.Update(ctx, warehouseID, req)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "not found")

		repo.AssertExpectations(t)
	})

	t.Run("name conflict", func(t *testing.T) {
		warehouseID := uuid.New()
		req := models.UpdateWarehouseRequest{
			Name: "Conflicting Name",
		}

		existingWarehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Original Name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(existingWarehouse, nil)
		repo.On("ExistsByName", ctx, "Conflicting Name", &warehouseID).Return(true, nil)

		response, err := service.Update(ctx, warehouseID, req)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "already exists")

		repo.AssertExpectations(t)
	})
}

func TestWarehouseService_Delete(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		warehouseID := uuid.New()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Test Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(warehouse, nil)
		repo.On("HasActiveEquipment", ctx, warehouseID).Return(false, nil)
		repo.On("SoftDelete", ctx, warehouseID).Return(nil)

		err := service.Delete(ctx, warehouseID)
		require.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("warehouse not found", func(t *testing.T) {
		warehouseID := uuid.New()

		repo.On("GetByID", ctx, warehouseID, false).Return(nil, nil)

		err := service.Delete(ctx, warehouseID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		repo.AssertExpectations(t)
	})

	t.Run("has active equipment", func(t *testing.T) {
		warehouseID := uuid.New()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Test Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetByID", ctx, warehouseID, false).Return(warehouse, nil)
		repo.On("HasActiveEquipment", ctx, warehouseID).Return(true, nil)

		err := service.Delete(ctx, warehouseID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "equipment is still present")

		repo.AssertExpectations(t)
	})
}

func TestWarehouseService_Restore(t *testing.T) {
	repo := &MockWarehouseRepository{}
	service := NewWarehouseService(repo)
	ctx := context.Background()

	t.Run("successful restoration", func(t *testing.T) {
		warehouseID := uuid.New()
		deletedTime := time.Now()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Deleted Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: &deletedTime,
		}

		restoredWarehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Deleted Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}

		repo.On("GetByID", ctx, warehouseID, true).Return(warehouse, nil)
		repo.On("ExistsByName", ctx, "Deleted Warehouse", &warehouseID).Return(false, nil)
		repo.On("Restore", ctx, warehouseID).Return(nil)
		repo.On("GetByID", ctx, warehouseID, false).Return(restoredWarehouse, nil)

		response, err := service.Restore(ctx, warehouseID)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Nil(t, response.DeletedAt)

		repo.AssertExpectations(t)
	})

	t.Run("warehouse not found", func(t *testing.T) {
		warehouseID := uuid.New()

		repo.On("GetByID", ctx, warehouseID, true).Return(nil, nil)

		response, err := service.Restore(ctx, warehouseID)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "not found")

		repo.AssertExpectations(t)
	})

	t.Run("not deleted", func(t *testing.T) {
		warehouseID := uuid.New()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Active Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}

		repo.On("GetByID", ctx, warehouseID, true).Return(warehouse, nil)

		response, err := service.Restore(ctx, warehouseID)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "not deleted")

		repo.AssertExpectations(t)
	})

	t.Run("name conflict on restore", func(t *testing.T) {
		warehouseID := uuid.New()
		deletedTime := time.Now()
		warehouse := &models.Warehouse{
			ID:        warehouseID,
			Name:      "Deleted Warehouse",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: &deletedTime,
		}

		repo.On("GetByID", ctx, warehouseID, true).Return(warehouse, nil)
		repo.On("ExistsByName", ctx, "Deleted Warehouse", &warehouseID).Return(true, nil)

		response, err := service.Restore(ctx, warehouseID)
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "conflicts with existing warehouse")

		repo.AssertExpectations(t)
	})
}
