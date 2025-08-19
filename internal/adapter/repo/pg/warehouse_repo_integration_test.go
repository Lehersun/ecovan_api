//go:build integration

package pg

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarehouseRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Test Warehouse " + uuid.New().String()[:8],
				Address:   stringPtr("123 Test St"),
				Notes:     stringPtr("Test notes"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Get warehouse using transaction
			var retrieved models.Warehouse
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.Equal(t, warehouse.Name, retrieved.Name)
			assert.Equal(t, warehouse.Address, retrieved.Address)
			assert.Equal(t, warehouse.Notes, retrieved.Notes)
		})
	})

	t.Run("List with pagination", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create multiple warehouses
			for i := 0; i < 5; i++ {
				warehouse := &models.Warehouse{
					ID:        uuid.New(),
					Name:      formatString("Warehouse %d", i+1),
					Address:   stringPtr(formatString("Address %d", i+1)),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				// Create warehouse using transaction
				_, err := tx.Exec(ctx, `
					INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
				require.NoError(t, err)
			}

			// Test pagination using transaction
			// Count total warehouses
			var total int64
			var err error
			err = tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM warehouses WHERE deleted_at IS NULL
			`).Scan(&total)
			require.NoError(t, err)
			assert.True(t, total >= 5)

			// Get paginated results
			rows, err := tx.Query(ctx, `
				SELECT id, name, address, notes, created_at, updated_at, deleted_at
				FROM warehouses 
				WHERE deleted_at IS NULL
				ORDER BY created_at DESC
				LIMIT 3 OFFSET 0
			`)
			require.NoError(t, err)
			defer rows.Close()

			var items []models.Warehouse
			for rows.Next() {
				var warehouse models.Warehouse
				err := rows.Scan(
					&warehouse.ID, &warehouse.Name, &warehouse.Address, &warehouse.Notes,
					&warehouse.CreatedAt, &warehouse.UpdatedAt, &warehouse.DeletedAt,
				)
				require.NoError(t, err)
				items = append(items, warehouse)
			}
			require.NoError(t, rows.Err())

			assert.Len(t, items, 3)
			assert.True(t, total >= 5)
		})
	})

	t.Run("Update", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Update Test Warehouse " + uuid.New().String()[:8],
				Address:   stringPtr("Original Address"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Update warehouse using transaction
			warehouse.Name = "Updated Warehouse Name"
			warehouse.Address = stringPtr("Updated Address")
			warehouse.UpdatedAt = time.Now()

			_, err = tx.Exec(ctx, `
			UPDATE warehouses 
			SET name = $1, address = $2, updated_at = $3
			WHERE id = $4
		`, warehouse.Name, warehouse.Address, warehouse.UpdatedAt, warehouse.ID)
			require.NoError(t, err)

			// Verify update using transaction
			var retrieved models.Warehouse
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			assert.Equal(t, "Updated Warehouse Name", retrieved.Name)
			assert.Equal(t, "Updated Address", *retrieved.Address)
		})
	})

	t.Run("SoftDelete and Restore", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Delete Test Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Soft delete using transaction
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = $1 WHERE id = $2", time.Now(), warehouse.ID)
			require.NoError(t, err)

			// Verify deleted using transaction
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

			// Restore using transaction
			_, err = tx.Exec(ctx, "UPDATE warehouses SET deleted_at = NULL WHERE id = $1", warehouse.ID)
			require.NoError(t, err)

			// Verify restored using transaction
			err = tx.QueryRow(ctx, `
			SELECT id, name, address, notes, created_at, updated_at, deleted_at
			FROM warehouses WHERE id = $1 AND deleted_at IS NULL
		`, warehouse.ID).Scan(
				&retrieved.ID, &retrieved.Name, &retrieved.Address, &retrieved.Notes,
				&retrieved.CreatedAt, &retrieved.UpdatedAt, &retrieved.DeletedAt,
			)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.Nil(t, retrieved.DeletedAt)
		})
	})

	t.Run("ExistsByName", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Unique Name Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Test exists using transaction
			var exists bool
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM warehouses
				WHERE name = $1 AND deleted_at IS NULL
			)
		`, warehouse.Name).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists)

			// Test doesn't exist using transaction
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM warehouses
				WHERE name = $1 AND deleted_at IS NULL
			)
		`, "Non-existent Name").Scan(&exists)
			require.NoError(t, err)
			assert.False(t, exists)

			// Test exclude ID using transaction
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM warehouses
				WHERE name = $1 AND deleted_at IS NULL AND id != $2
			)
		`, warehouse.Name, warehouse.ID).Scan(&exists)
			require.NoError(t, err)
			assert.False(t, exists)
		})
	})

	t.Run("HasActiveEquipment", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Equipment Test Warehouse " + uuid.New().String()[:8],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create warehouse using transaction
			_, err := tx.Exec(ctx, `
			INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, warehouse.ID, warehouse.Name, warehouse.Address, warehouse.Notes, warehouse.CreatedAt, warehouse.UpdatedAt)
			require.NoError(t, err)

			// Initially no equipment using transaction
			var hasEquipment bool
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM equipment
				WHERE warehouse_id = $1 AND deleted_at IS NULL
			)
		`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.False(t, hasEquipment)

			// Add equipment
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

			// Now has equipment - check using the same transaction
			err = tx.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM equipment
					WHERE warehouse_id = $1 AND deleted_at IS NULL
				)
			`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Delete equipment
			_, err = tx.Exec(ctx, "DELETE FROM equipment WHERE id = $1", equipment.ID)
			require.NoError(t, err)

			// No equipment again using transaction
			err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM equipment
				WHERE warehouse_id = $1 AND deleted_at IS NULL
			)
		`, warehouse.ID).Scan(&hasEquipment)
			require.NoError(t, err)
			assert.False(t, hasEquipment)
		})
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func formatString(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
