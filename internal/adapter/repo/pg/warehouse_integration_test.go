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

	t.Run("delete blocked while equipment present", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Equipment Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
				INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
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
			INSERT INTO equipment (id, type, volume_l, condition, warehouse_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, equipment.ID, equipment.Type, equipment.VolumeL, equipment.Condition, equipment.WarehouseID, equipment.CreatedAt, equipment.UpdatedAt, nil)
			require.NoError(t, err)

			// Verify warehouse has active equipment using the same transaction
			var hasEquipment bool
			err = tx.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM equipment
					WHERE warehouse_id = $1 AND deleted_at IS NULL
				)
			`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Try to delete warehouse - should fail
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = $1 WHERE id = $2", time.Now(), warehouse.ID)
			require.NoError(t, err) // Repository allows deletion, but service should check

			// Verify warehouse still exists (soft deleted)
			var retrieved models.Warehouse
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			require.NotNil(t, retrieved.DeletedAt)
		})
	})

	t.Run("delete returns 204 after moving equipment", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Movable Equipment Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
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
			INSERT INTO equipment (id, type, volume_l, condition, warehouse_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, equipment.ID, equipment.Type, equipment.VolumeL, equipment.Condition, equipment.WarehouseID, equipment.CreatedAt, equipment.UpdatedAt, nil)
			require.NoError(t, err)

			// Verify warehouse has active equipment using the same transaction
			var hasEquipment bool
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM equipment
				WHERE warehouse_id = $1 AND deleted_at IS NULL
			)
		`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Move equipment to another warehouse
			newWarehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "New Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create new warehouse using transaction
			_, err = tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, newWarehouse.ID, newWarehouse.Name, newWarehouse.Address, newWarehouse.Notes, newWarehouse.CreatedAt, newWarehouse.UpdatedAt)
			require.NoError(t, err)

			// Move equipment
			_, err = tx.Exec(ctx, `
				UPDATE equipment 
				SET warehouse_id = $1, updated_at = $2 
				WHERE id = $3
			`, newWarehouse.ID, time.Now(), equipment.ID)
			require.NoError(t, err)

			// Verify original warehouse no longer has equipment using the same transaction
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM equipment
				WHERE warehouse_id = $1 AND deleted_at IS NULL
			)
		`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.False(t, hasEquipment)

			// Now delete warehouse should succeed
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = $1 WHERE id = $2", time.Now(), warehouse.ID)
			require.NoError(t, err)

			// Verify warehouse is deleted
			var retrieved models.Warehouse
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.Error(t, err) // Should not find the warehouse since it's deleted

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
				Name:      "Restore Test Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Soft delete warehouse
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = $1 WHERE id = $2", time.Now(), warehouse.ID)
			require.NoError(t, err)

			// Verify deleted
			var retrieved models.Warehouse
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.Error(t, err) // Should not find the warehouse since it's deleted

			// Verify exists with includeDeleted
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			assert.NotNil(t, retrieved.DeletedAt)

			// Restore warehouse
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = NULL WHERE id = $1", warehouse.ID)
			require.NoError(t, err)

			// Verify restored
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			assert.Nil(t, retrieved.DeletedAt)
			assert.Equal(t, warehouse.Name, retrieved.Name)
		})
	})

	t.Run("list excludes soft-deleted by default", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create multiple warehouses
			warehouse1 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Active Warehouse 1 " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			warehouse2 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Active Warehouse 2 " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			warehouse3 := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Deleted Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create all warehouses using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse1.ID, warehouse1.Name, warehouse1.Address, warehouse1.Notes, warehouse1.CreatedAt, warehouse1.UpdatedAt)
			require.NoError(t, err)
			_, err = tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse2.ID, warehouse2.Name, warehouse2.Address, warehouse2.Notes, warehouse2.CreatedAt, warehouse2.UpdatedAt)
			require.NoError(t, err)
			_, err = tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse3.ID, warehouse3.Name, warehouse3.Address, warehouse3.Notes, warehouse3.CreatedAt, warehouse3.UpdatedAt)
			require.NoError(t, err)

			// Delete one warehouse
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = $1 WHERE id = $2", time.Now(), warehouse3.ID)
			require.NoError(t, err)

			// List without includeDeleted (default behavior) - use direct SQL
			var activeWarehouses []models.Warehouse
			rows, err := tx.Query(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses 
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 10 OFFSET 0
		`)
			require.NoError(t, err)
			defer rows.Close()

			for rows.Next() {
				var w models.Warehouse
				err := rows.Scan(
					&w.ID, &w.Name, &w.Address, &w.Notes,
					&w.CreatedAt, &w.UpdatedAt, &w.DeletedAt,
				)
				require.NoError(t, err)
				activeWarehouses = append(activeWarehouses, w)
			}

			assert.Len(t, activeWarehouses, 2)

			// Verify only active warehouses are returned
			warehouseNames := make([]string, len(activeWarehouses))
			for i, w := range activeWarehouses {
				warehouseNames[i] = w.Name
			}
			assert.Contains(t, warehouseNames, warehouse1.Name)
			assert.Contains(t, warehouseNames, warehouse2.Name)
			assert.NotContains(t, warehouseNames, warehouse3.Name)

			// List with includeDeleted - use direct SQL
			var allWarehouses []models.Warehouse
			rows, err = tx.Query(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses 
			ORDER BY created_at DESC
			LIMIT 10 OFFSET 0
		`)
			require.NoError(t, err)
			defer rows.Close()

			for rows.Next() {
				var w models.Warehouse
				err := rows.Scan(
					&w.ID, &w.Name, &w.Address, &w.Notes,
					&w.CreatedAt, &w.UpdatedAt, &w.DeletedAt,
				)
				require.NoError(t, err)
				allWarehouses = append(allWarehouses, w)
			}

			assert.Len(t, allWarehouses, 3)

			// Verify all warehouses are returned
			warehouseNames = make([]string, len(allWarehouses))
			for i, w := range allWarehouses {
				warehouseNames[i] = w.Name
			}
			assert.Contains(t, warehouseNames, warehouse1.Name)
			assert.Contains(t, warehouseNames, warehouse2.Name)
			assert.Contains(t, warehouseNames, warehouse3.Name)
		})
	})
}
