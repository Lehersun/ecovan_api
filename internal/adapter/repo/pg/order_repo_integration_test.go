//go:build integration

package pg

import (
	"context"
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create repositories
	orderRepo := NewOrderRepository(TestPool)

	t.Run("Create and retrieve order", func(t *testing.T) {
		ctx := context.Background()

		// Create test client and object using the helper functions with unique names
		clientID := MakeClient(t, ctx, TestPool, "OrderTest-CreateRetrieve-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "OrderTest-Office-CreateRetrieve-"+uuid.New().String()[:8])

		// Create test order using the repository
		order := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:        "DRAFT",
			Priority:      "MEDIUM",
			Notes:         stringPtrOrder("Test order"),
		}

		err := orderRepo.Create(ctx, &order)
		require.NoError(t, err)

		// Test GetByID using the repository
		retrievedOrder, err := orderRepo.GetByID(ctx, order.ID, false)
		require.NoError(t, err)
		require.NotNil(t, retrievedOrder)

		// Verify retrieved order
		assert.Equal(t, order.ID, retrievedOrder.ID)
		assert.Equal(t, clientID, retrievedOrder.ClientID)
		assert.Equal(t, objectID, retrievedOrder.ObjectID)
		assert.Equal(t, order.ScheduledDate, retrievedOrder.ScheduledDate)
		assert.Equal(t, "DRAFT", retrievedOrder.Status)
		assert.Equal(t, "Test order", *retrievedOrder.Notes)
		assert.Nil(t, retrievedOrder.DeletedAt)
	})

	t.Run("List orders", func(t *testing.T) {
		ctx := context.Background()

		// Create test client and object with unique names
		clientID := MakeClient(t, ctx, TestPool, "OrderTest-ListOrders-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "OrderTest-Office-ListOrders-"+uuid.New().String()[:8])

		// Create test orders using the repository
		order1 := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:        "DRAFT",
			Priority:      "MEDIUM",
			Notes:         stringPtrOrder("First order"),
		}
		order2 := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			Status:        "SCHEDULED",
			Priority:      "HIGH",
			Notes:         stringPtrOrder("Second order"),
		}

		err := orderRepo.Create(ctx, &order1)
		require.NoError(t, err)
		err = orderRepo.Create(ctx, &order2)
		require.NoError(t, err)

		// Test List using the repository
		req := models.OrderListRequest{
			Page:           1,
			PageSize:       20,
			IncludeDeleted: false,
		}

		result, err := orderRepo.List(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have at least 2 orders from our test
		assert.GreaterOrEqual(t, len(result.Items), 2)
		assert.GreaterOrEqual(t, result.Total, int64(2))
		assert.Equal(t, int64(1), int64(result.Page))
		assert.Equal(t, int64(20), int64(result.PageSize))
	})

	t.Run("List orders with filters", func(t *testing.T) {
		ctx := context.Background()

		// Create test client and object with unique names
		clientID := MakeClient(t, ctx, TestPool, "OrderTest-ListFilters-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "OrderTest-Office-ListFilters-"+uuid.New().String()[:8])

		// Create test order using the repository
		order := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:        "DRAFT",
			Priority:      "LOW",
		}

		err := orderRepo.Create(ctx, &order)
		require.NoError(t, err)

		// Test filter by client
		req := models.OrderListRequest{
			Page:           1,
			PageSize:       20,
			ClientID:       &clientID,
			IncludeDeleted: false,
		}

		result, err := orderRepo.List(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)

		// All orders should belong to the test client
		for _, orderResp := range result.Items {
			assert.Equal(t, clientID, orderResp.ClientID)
		}
	})

	t.Run("Soft delete and restore", func(t *testing.T) {
		ctx := context.Background()

		// Create test client and object with unique names
		clientID := MakeClient(t, ctx, TestPool, "OrderTest-SoftDelete-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "OrderTest-Office-SoftDelete-"+uuid.New().String()[:8])

		// Create test order using the repository
		order := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:        "DRAFT",
			Priority:      "MEDIUM",
		}

		err := orderRepo.Create(ctx, &order)
		require.NoError(t, err)

		// Test soft delete
		err = orderRepo.SoftDelete(ctx, order.ID)
		require.NoError(t, err)

		// Should not be found with includeDeleted=false
		retrievedOrder, err := orderRepo.GetByID(ctx, order.ID, false)
		require.NoError(t, err)
		assert.Nil(t, retrievedOrder)

		// Should be found with includeDeleted=true
		retrievedOrder, err = orderRepo.GetByID(ctx, order.ID, true)
		require.NoError(t, err)
		require.NotNil(t, retrievedOrder)
		assert.NotNil(t, retrievedOrder.DeletedAt)

		// Test restore
		err = orderRepo.Restore(ctx, order.ID)
		require.NoError(t, err)

		// Should be found again with includeDeleted=false
		retrievedOrder, err = orderRepo.GetByID(ctx, order.ID, false)
		require.NoError(t, err)
		require.NotNil(t, retrievedOrder)
		assert.Nil(t, retrievedOrder.DeletedAt)
	})

	t.Run("Update order", func(t *testing.T) {
		ctx := context.Background()

		// Create test client and object with unique names
		clientID := MakeClient(t, ctx, TestPool, "OrderTest-UpdateOrder-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "OrderTest-Office-UpdateOrder-"+uuid.New().String()[:8])

		// Create test order
		order := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
			Status:        string(models.OrderStatusDraft),
			Priority:      "HIGH",
			Notes:         stringPtrOrder("Order to update"),
		}

		err := orderRepo.Create(ctx, &order)
		require.NoError(t, err)

		// Update the order
		order.Status = string(models.OrderStatusScheduled)
		order.Notes = stringPtrOrder("Updated order")

		err = orderRepo.Update(ctx, &order)
		require.NoError(t, err)

		// Retrieve and verify update
		retrievedOrder, err := orderRepo.GetByID(ctx, order.ID, false)
		require.NoError(t, err)
		require.NotNil(t, retrievedOrder)

		assert.Equal(t, string(models.OrderStatusScheduled), retrievedOrder.Status)
		assert.Equal(t, "Updated order", *retrievedOrder.Notes)
	})
}

func TestOrderRepository_ExistsByClientAndObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	orderRepo := NewOrderRepository(TestPool)

	t.Run("No existing order", func(t *testing.T) {
		ctx := context.Background()

		clientID := MakeClient(t, ctx, TestPool, "OrderTest-NoOrders-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "MainOffice-NoOrders-"+uuid.New().String()[:8])

		exists, err := orderRepo.ExistsByClientAndObject(ctx, clientID, objectID, nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("With existing order", func(t *testing.T) {
		ctx := context.Background()

		clientID := MakeClient(t, ctx, TestPool, "OrderTest-WithOrders-"+uuid.New().String()[:8])
		objectID := MakeClientObject(t, ctx, TestPool, clientID, "MainOffice-WithOrders-"+uuid.New().String()[:8])

		// Create test order using the repository
		order := models.Order{
			ClientID:      clientID,
			ObjectID:      objectID,
			ScheduledDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:        "DRAFT",
			Priority:      "MEDIUM",
		}

		err := orderRepo.Create(ctx, &order)
		require.NoError(t, err)

		// Should exist
		exists, err := orderRepo.ExistsByClientAndObject(ctx, clientID, objectID, nil)
		require.NoError(t, err)
		assert.True(t, exists)

		// Should not exist when excluding this order
		exists, err = orderRepo.ExistsByClientAndObject(ctx, clientID, objectID, &order.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// Helper function to create string pointer for orders
func stringPtrOrder(s string) *string {
	return &s
}
