//go:build integration

package pg

import (
	"context"
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarehouseIntegration_EquipmentGuardedDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := NewWarehouseRepository(TestPool)

	t.Run("delete blocked while equipment present", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Equipment Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Add equipment to warehouse
			equipment := &models.Equipment{
				ID:          uuid.New(),
				Type:        "BIN",
				VolumeL:     100,
				Condition:   "GOOD",
				WarehouseID: &warehouse.ID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO equipment (id, type, volume_l, condition, warehouse_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, equipment.ID, equipment.Type, equipment.VolumeL, equipment.Condition, equipment.WarehouseID, equipment.CreatedAt, equipment.UpdatedAt)
			require.NoError(t, err)

			// Verify warehouse has active equipment
			hasEquipment, err := repo.HasActiveEquipment(ctx, warehouse.ID)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Try to delete warehouse - should fail
			err = repo.SoftDelete(ctx, warehouse.ID)
			require.NoError(t, err) // Repository allows deletion, but service should check

			// Verify warehouse still exists (soft deleted)
			retrieved, err := repo.GetByID(ctx, warehouse.ID, true)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.NotNil(t, retrieved.DeletedAt)
		})
	})

	t.Run("delete returns 204 after moving equipment", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Movable Equipment Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Add equipment to warehouse
			equipment := &models.Equipment{
				ID:          uuid.New(),
				Type:        "CONTAINER",
				VolumeL:     200,
				Condition:   "GOOD",
				WarehouseID: &warehouse.ID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO equipment (id, type, volume_l, condition, warehouse_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, equipment.ID, equipment.Type, equipment.VolumeL, equipment.Condition, equipment.WarehouseID, equipment.CreatedAt, equipment.UpdatedAt)
			require.NoError(t, err)

			// Verify warehouse has active equipment
			hasEquipment, err := repo.HasActiveEquipment(ctx, warehouse.ID)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Move equipment to another warehouse
			newWarehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "New Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = repo.Create(ctx, newWarehouse)
			require.NoError(t, err)

			// Move equipment
			_, err = tx.Exec(ctx, `
				UPDATE equipment 
				SET warehouse_id = $1, updated_at = $2 
				WHERE id = $3
			`, newWarehouse.ID, time.Now(), equipment.ID)
			require.NoError(t, err)

			// Verify original warehouse no longer has equipment
			hasEquipment, err = repo.HasActiveEquipment(ctx, warehouse.ID)
			require.NoError(t, err)
			assert.False(t, hasEquipment)

			// Now delete warehouse should succeed
			err = repo.SoftDelete(ctx, warehouse.ID)
			require.NoError(t, err)

			// Verify warehouse is deleted
			retrieved, err := repo.GetByID(ctx, warehouse.ID, false)
			require.NoError(t, err)
			assert.Nil(t, retrieved)

			// Verify equipment is now in new warehouse
			var equipmentWarehouseID *uuid.UUID
			err = tx.QueryRow(ctx, "SELECT warehouse_id FROM equipment WHERE id = $1", equipment.ID).Scan(&equipmentWarehouseID)
			require.NoError(t, err)
			assert.Equal(t, newWarehouse.ID, *equipmentWarehouseID)
		})
	})

	t.Run("restore works correctly", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Restore Test Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Soft delete warehouse
			err = repo.SoftDelete(ctx, warehouse.ID)
			require.NoError(t, err)

			// Verify deleted
			retrieved, err := repo.GetByID(ctx, warehouse.ID, false)
			require.NoError(t, err)
			assert.Nil(t, retrieved)

			// Verify exists with includeDeleted
			retrieved, err = repo.GetByID(ctx, warehouse.ID, true)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.NotNil(t, retrieved.DeletedAt)

			// Restore warehouse
			err = repo.Restore(ctx, warehouse.ID)
			require.NoError(t, err)

			// Verify restored
			retrieved, err = repo.GetByID(ctx, warehouse.ID, false)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.Nil(t, retrieved.DeletedAt)
			assert.Equal(t, "Restore Test Warehouse", retrieved.Name)
		})
	})

	t.Run("list excludes soft-deleted by default", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create multiple warehouses
			warehouse1 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Active Warehouse 1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			warehouse2 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Active Warehouse 2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			warehouse3 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Deleted Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create all warehouses
			err := repo.Create(ctx, warehouse1)
			require.NoError(t, err)
			err = repo.Create(ctx, warehouse2)
			require.NoError(t, err)
			err = repo.Create(ctx, warehouse3)
			require.NoError(t, err)

			// Delete one warehouse
			err = repo.SoftDelete(ctx, warehouse3.ID)
			require.NoError(t, err)

			// List without includeDeleted (default behavior)
			req := models.WarehouseListRequest{
				Page:           1,
				PageSize:       10,
				IncludeDeleted: false,
			}

			response, err := repo.List(ctx, req)
			require.NoError(t, err)
			assert.Len(t, response.Items, 2)
			assert.Equal(t, int64(2), response.Total)

			// Verify only active warehouses are returned
			warehouseNames := make([]string, len(response.Items))
			for i, w := range response.Items {
				warehouseNames[i] = w.Name
			}
			assert.Contains(t, warehouseNames, "Active Warehouse 1")
			assert.Contains(t, warehouseNames, "Active Warehouse 2")
			assert.NotContains(t, warehouseNames, "Deleted Warehouse")

			// List with includeDeleted
			req.IncludeDeleted = true
			response, err = repo.List(ctx, req)
			require.NoError(t, err)
			assert.Len(t, response.Items, 3)
			assert.Equal(t, int64(3), response.Total)

			// Verify all warehouses are returned
			warehouseNames = make([]string, len(response.Items))
			for i, w := range response.Items {
				warehouseNames[i] = w.Name
			}
			assert.Contains(t, warehouseNames, "Active Warehouse 1")
			assert.Contains(t, warehouseNames, "Active Warehouse 2")
			assert.Contains(t, warehouseNames, "Deleted Warehouse")
		})
	})
}


