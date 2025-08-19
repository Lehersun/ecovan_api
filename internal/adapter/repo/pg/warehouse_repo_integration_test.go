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

	repo := NewWarehouseRepository(TestPool)

	t.Run("Create and GetByID", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Test Warehouse",
				Address:   stringPtr("123 Test St"),
				Notes:     stringPtr("Test notes"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Get warehouse
			retrieved, err := repo.GetByID(ctx, warehouse.ID, false)
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
				err := repo.Create(ctx, warehouse)
				require.NoError(t, err)
			}

			// Test pagination
			req := models.WarehouseListRequest{
				Page:           1,
				PageSize:       3,
				IncludeDeleted: false,
			}

			response, err := repo.List(ctx, req)
			require.NoError(t, err)
			assert.Len(t, response.Items, 3)
			assert.Equal(t, 1, response.Page)
			assert.Equal(t, 3, response.PageSize)
			assert.True(t, response.Total >= 5)
		})
	})

	t.Run("Update", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Update Test Warehouse",
				Address:   stringPtr("Original Address"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Update warehouse
			warehouse.Name = "Updated Warehouse Name"
			warehouse.Address = stringPtr("Updated Address")
			warehouse.UpdatedAt = time.Now()

			err = repo.Update(ctx, warehouse)
			require.NoError(t, err)

			// Verify update
			retrieved, err := repo.GetByID(ctx, warehouse.ID, false)
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
				Name:      "Delete Test Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Soft delete
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

			// Restore
			err = repo.Restore(ctx, warehouse.ID)
			require.NoError(t, err)

			// Verify restored
			retrieved, err = repo.GetByID(ctx, warehouse.ID, false)
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
				Name:      "Unique Name Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Test exists
			exists, err := repo.ExistsByName(ctx, "Unique Name Warehouse", nil)
			require.NoError(t, err)
			assert.True(t, exists)

			// Test doesn't exist
			exists, err = repo.ExistsByName(ctx, "Non-existent Name", nil)
			require.NoError(t, err)
			assert.False(t, exists)

			// Test exclude ID
			exists, err = repo.ExistsByName(ctx, "Unique Name Warehouse", &warehouse.ID)
			require.NoError(t, err)
			assert.False(t, exists)
		})
	})

	t.Run("HasActiveEquipment", func(t *testing.T) {
		WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			// Create warehouse
			warehouse := &models.Warehouse{
				ID:        uuid.New(),
				Name:      "Equipment Test Warehouse",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := repo.Create(ctx, warehouse)
			require.NoError(t, err)

			// Initially no equipment
			hasEquipment, err := repo.HasActiveEquipment(ctx, warehouse.ID)
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
				INSERT INTO equipment (id, type, volume_l, condition, warehouse_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, equipment.ID, equipment.Type, equipment.VolumeL, equipment.Condition, equipment.WarehouseID, equipment.CreatedAt, equipment.UpdatedAt)
			require.NoError(t, err)

			// Now has equipment
			hasEquipment, err = repo.HasActiveEquipment(ctx, warehouse.ID)
			require.NoError(t, err)
			assert.True(t, hasEquipment)

			// Delete equipment
			_, err = tx.Exec(ctx, "DELETE FROM equipment WHERE id = $1", equipment.ID)
			require.NoError(t, err)

			// No equipment again
			hasEquipment, err = repo.HasActiveEquipment(ctx, warehouse.ID)
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
