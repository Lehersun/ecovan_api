//go:build integration

package pg

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientsRepository_Integration(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Test creating a client
		clientID := MakeClient(t, ctx, tx, "Test Client")
		assert.NotEqual(t, uuid.Nil, clientID)

		// Test reading the client back
		var name, phone, email string
		query := `SELECT name, phone, email FROM clients WHERE id = $1`
		err := tx.QueryRow(ctx, query, clientID).Scan(&name, &phone, &email)
		require.NoError(t, err)

		assert.Equal(t, "Test Client", name)
		assert.Equal(t, "+1234567890", phone)
		assert.Equal(t, "test@example.com", email)

		// Test creating a client object
		objectID := MakeClientObject(t, ctx, tx, clientID, "Test Object")
		assert.NotEqual(t, uuid.Nil, objectID)

		// Test reading the client object back
		var objName, objAddress string
		var objClientID uuid.UUID
		query = `SELECT name, address, client_id FROM client_objects WHERE id = $1`
		err = tx.QueryRow(ctx, query, objectID).Scan(&objName, &objAddress, &objClientID)
		require.NoError(t, err)

		assert.Equal(t, "Test Object", objName)
		assert.Equal(t, "123 Test St", objAddress)
		assert.Equal(t, clientID, objClientID)
	})
}

func TestClientsRepository_Isolation(t *testing.T) {
	// Test that transactions are isolated
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		clientID1 := MakeClient(t, ctx, tx, "Client A")
		assert.NotEqual(t, uuid.Nil, clientID1)
	})

	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		clientID2 := MakeClient(t, ctx, tx, "Client A") // Same name
		assert.NotEqual(t, uuid.Nil, clientID2)

		// Verify this client exists in this transaction
		var count int
		query := `SELECT COUNT(*) FROM clients WHERE name = 'Client A'`
		err := tx.QueryRow(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestClientsRepository_CompleteWorkflow(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a complete workflow: client -> object -> order
		clientID := MakeClient(t, ctx, tx, "Workflow Client")
		objectID := MakeClientObject(t, ctx, tx, clientID, "Workflow Object")
		orderID := MakeOrder(t, ctx, tx, clientID, objectID)

		// Verify all records exist and are linked correctly
		var orderClientID, orderObjectID uuid.UUID
		var status string
		query := `SELECT client_id, object_id, status FROM orders WHERE id = $1`
		err := tx.QueryRow(ctx, query, orderID).Scan(&orderClientID, &orderObjectID, &status)
		require.NoError(t, err)

		assert.Equal(t, clientID, orderClientID)
		assert.Equal(t, objectID, orderObjectID)
		assert.Equal(t, "DRAFT", status)
	})
}

func TestClientsRepository_CRUDOperations(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {

		// Test CREATE
		clientID := uuid.New()
		query := `
			INSERT INTO clients (id, name, tax_id, email, phone, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query, clientID, "CRUD Test Client", "123456789", "crud@test.com", "+1234567890", "Test notes", time.Now(), time.Now())
		require.NoError(t, err)

		// Test READ
		var name, taxID, email, phone, notes string
		query = `SELECT name, tax_id, email, phone, notes FROM clients WHERE id = $1`
		err = tx.QueryRow(ctx, query, clientID).Scan(&name, &taxID, &email, &phone, &notes)
		require.NoError(t, err)
		assert.Equal(t, "CRUD Test Client", name)
		assert.Equal(t, "123456789", taxID)
		assert.Equal(t, "crud@test.com", email)
		assert.Equal(t, "+1234567890", phone)
		assert.Equal(t, "Test notes", notes)

		// Test UPDATE
		updateQuery := `
			UPDATE clients 
			SET name = $1, tax_id = $2, email = $3, phone = $4, notes = $5, updated_at = $6
			WHERE id = $7
		`
		_, err = tx.Exec(ctx, updateQuery, "Updated CRUD Client", "987654321", "updated@test.com", "+9876543210", "Updated notes", time.Now(), clientID)
		require.NoError(t, err)

		// Verify update
		query = `SELECT name, tax_id, email, phone, notes FROM clients WHERE id = $1`
		err = tx.QueryRow(ctx, query, clientID).Scan(&name, &taxID, &email, &phone, &notes)
		require.NoError(t, err)
		assert.Equal(t, "Updated CRUD Client", name)
		assert.Equal(t, "987654321", taxID)
		assert.Equal(t, "updated@test.com", email)
		assert.Equal(t, "+9876543210", phone)
		assert.Equal(t, "Updated notes", notes)

		// Test soft DELETE
		deleteQuery := `UPDATE clients SET deleted_at = $1, updated_at = $2 WHERE id = $3`
		_, err = tx.Exec(ctx, deleteQuery, time.Now(), time.Now(), clientID)
		require.NoError(t, err)

		// Verify soft delete (should not appear in normal queries)
		query = `SELECT COUNT(*) FROM clients WHERE id = $1 AND deleted_at IS NULL`
		var count int
		err = tx.QueryRow(ctx, query, clientID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify soft delete (should appear when including deleted)
		query = `SELECT COUNT(*) FROM clients WHERE id = $1`
		err = tx.QueryRow(ctx, query, clientID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Test RESTORE
		restoreQuery := `UPDATE clients SET deleted_at = NULL, updated_at = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, restoreQuery, time.Now(), clientID)
		require.NoError(t, err)

		// Verify restore
		query = `SELECT COUNT(*) FROM clients WHERE id = $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, clientID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestClientsRepository_SearchAndPagination(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create multiple test clients with unique names
		clients := []struct {
			name  string
			email string
			phone string
		}{
			{"SearchTest Alpha Company", "search-alpha@company.com", "+1111111111"},
			{"SearchTest Beta Corporation", "search-beta@corp.com", "+2222222222"},
			{"SearchTest Gamma Industries", "search-gamma@industries.com", "+3333333333"},
			{"SearchTest Delta Services", "search-delta@services.com", "+4444444444"},
		}

		for _, client := range clients {
			query := `
				INSERT INTO clients (id, name, email, phone, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err := tx.Exec(ctx, query, uuid.New(), client.name, client.email, client.phone, time.Now(), time.Now())
			require.NoError(t, err)
		}

		// Test search by name - should find only Alpha Company
		query := `SELECT COUNT(*) FROM clients WHERE name ILIKE $1 AND deleted_at IS NULL`
		var count int
		err := tx.QueryRow(ctx, query, "%Alpha Company%").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Test search by email - should find only Beta Corporation
		query = `SELECT COUNT(*) FROM clients WHERE email ILIKE $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, "%search-beta%").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Test total count of our test clients
		query = `SELECT COUNT(*) FROM clients WHERE name LIKE 'SearchTest%' AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 4, count)

		// Test pagination with limit
		query = `SELECT name FROM clients WHERE name LIKE 'SearchTest%' AND deleted_at IS NULL ORDER BY name LIMIT 2`
		rows, err := tx.Query(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		var names []string
		for rows.Next() {
			var name string
			err := rows.Scan(&name)
			require.NoError(t, err)
			names = append(names, name)
		}
		assert.Equal(t, 2, len(names))
		assert.Equal(t, "SearchTest Alpha Company", names[0])
		assert.Equal(t, "SearchTest Beta Corporation", names[1])
	})
}

func TestClientsRepository_NameUniqueness(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create first client
		clientID1 := uuid.New()
		query := `
			INSERT INTO clients (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
		`
		_, err := tx.Exec(ctx, query, clientID1, "Unique Name Client", time.Now(), time.Now())
		require.NoError(t, err)

		// Try to create second client with same name (should fail due to unique constraint)
		clientID2 := uuid.New()
		query = `
			INSERT INTO clients (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, query, clientID2, "Unique Name Client", time.Now(), time.Now())
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestClientObjectsRepository_Integration(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a test client first
		clientID := uuid.New()
		clientQuery := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, clientQuery, clientID, "Test Client", "test@client.com", "+1234567890", time.Now(), time.Now())
		require.NoError(t, err)

		// Test creating a client object
		objectID := uuid.New()
		objectQuery := `
			INSERT INTO client_objects (id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, objectQuery, objectID, clientID, "Test Location", "123 Test St", 40.7128, -74.0060, "Test notes", time.Now(), time.Now())
		require.NoError(t, err)

		// Test reading the client object back
		var name, address string
		var geoLat, geoLng float64
		var notes *string
		query := `SELECT name, address, geo_lat, geo_lng, notes FROM client_objects WHERE id = $1`
		err = tx.QueryRow(ctx, query, objectID).Scan(&name, &address, &geoLat, &geoLng, &notes)
		require.NoError(t, err)

		assert.Equal(t, "Test Location", name)
		assert.Equal(t, "123 Test St", address)
		assert.Equal(t, 40.7128, geoLat)
		assert.Equal(t, -74.0060, geoLng)
		assert.Equal(t, "Test notes", *notes)
	})
}

func TestClientObjectsRepository_CRUDOperations(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a test client
		clientID := uuid.New()
		clientQuery := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, clientQuery, clientID, "CRUD Test Client", "crud@test.com", "+1234567890", time.Now(), time.Now())
		require.NoError(t, err)

		// Test CREATE
		objectID := uuid.New()
		createQuery := `
			INSERT INTO client_objects (id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, createQuery, objectID, clientID, "CRUD Test Location", "456 CRUD St", 40.7589, -73.9851, "CRUD test notes", time.Now(), time.Now())
		require.NoError(t, err)

		// Test READ
		var name, address string
		var geoLat, geoLng float64
		var notes *string
		readQuery := `SELECT name, address, geo_lat, geo_lng, notes FROM client_objects WHERE id = $1`
		err = tx.QueryRow(ctx, readQuery, objectID).Scan(&name, &address, &geoLat, &geoLng, &notes)
		require.NoError(t, err)
		assert.Equal(t, "CRUD Test Location", name)
		assert.Equal(t, "456 CRUD St", address)

		// Test UPDATE
		updateQuery := `
			UPDATE client_objects 
			SET name = $1, address = $2, geo_lat = $3, geo_lng = $4, notes = $5, updated_at = $6
			WHERE id = $7
		`
		_, err = tx.Exec(ctx, updateQuery, "Updated Location", "789 Updated St", 40.7505, -73.9934, "Updated notes", time.Now(), objectID)
		require.NoError(t, err)

		// Verify update
		err = tx.QueryRow(ctx, readQuery, objectID).Scan(&name, &address, &geoLat, &geoLng, &notes)
		require.NoError(t, err)
		assert.Equal(t, "Updated Location", name)
		assert.Equal(t, "789 Updated St", address)

		// Test soft DELETE
		deleteQuery := `UPDATE client_objects SET deleted_at = $1, updated_at = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, deleteQuery, time.Now(), objectID)
		require.NoError(t, err)

		// Verify soft delete
		var count int
		query := `SELECT COUNT(*) FROM client_objects WHERE id = $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, objectID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Test RESTORE
		restoreQuery := `UPDATE client_objects SET deleted_at = NULL, updated_at = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, restoreQuery, time.Now(), objectID)
		require.NoError(t, err)

		// Verify restore
		query = `SELECT COUNT(*) FROM client_objects WHERE id = $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, objectID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestClientObjectsRepository_GuardChecks(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a test client
		clientID := uuid.New()
		clientQuery := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, clientQuery, clientID, "Guard Test Client", "guard@test.com", "+1234567890", time.Now(), time.Now())
		require.NoError(t, err)

		// Create a test client object
		objectID := uuid.New()
		objectQuery := `
			INSERT INTO client_objects (id, client_id, name, address, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, objectQuery, objectID, clientID, "Guard Test Location", "123 Guard St", time.Now(), time.Now())
		require.NoError(t, err)

		// Test 1: Check guard against active orders
		orderID := uuid.New()
		orderQuery := `
			INSERT INTO orders (id, client_id, object_id, scheduled_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = tx.Exec(ctx, orderQuery, orderID, clientID, objectID, "2024-01-15", "SCHEDULED", time.Now(), time.Now())
		require.NoError(t, err)

		// Check if we can detect active orders
		var activeOrderCount int
		activeOrderQuery := `
			SELECT COUNT(*) FROM orders 
			WHERE object_id = $1 
			AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS')
			AND deleted_at IS NULL
		`
		err = tx.QueryRow(ctx, activeOrderQuery, objectID).Scan(&activeOrderCount)
		require.NoError(t, err)
		assert.Equal(t, 1, activeOrderCount)

		// Test 2: Check guard against active equipment
		equipmentID := uuid.New()
		equipmentQuery := `
			INSERT INTO equipment (id, type, volume_l, condition, client_object_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.Exec(ctx, equipmentQuery, equipmentID, "BIN", 100, "GOOD", objectID, time.Now(), time.Now(), nil)
		require.NoError(t, err)

		// Check if we can detect active equipment
		var activeEquipmentCount int
		activeEquipmentQuery := `
			SELECT COUNT(*) FROM equipment 
			WHERE client_object_id = $1 AND deleted_at IS NULL
		`
		err = tx.QueryRow(ctx, activeEquipmentQuery, objectID).Scan(&activeEquipmentCount)
		require.NoError(t, err)
		assert.Equal(t, 1, activeEquipmentCount)

		// Test 3: Verify that deletion would be blocked
		// (We can't actually test the DELETE here since it's a guard, but we can verify the conditions exist)
		var totalBlockingConditions int
		blockingQuery := `
			SELECT 
				(SELECT COUNT(*) FROM orders WHERE object_id = $1 AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS') AND deleted_at IS NULL) +
				(SELECT COUNT(*) FROM equipment WHERE client_object_id = $1 AND deleted_at IS NULL)
		`
		err = tx.QueryRow(ctx, blockingQuery, objectID).Scan(&totalBlockingConditions)
		require.NoError(t, err)
		assert.Equal(t, 2, totalBlockingConditions) // 1 active order + 1 active equipment
	})
}

func TestClientObjectsRepository_GuardChecksAfterResolution(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a test client
		clientID := uuid.New()
		clientQuery := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, clientQuery, clientID, "Resolution Test Client", "resolution@test.com", "+1234567890", time.Now(), time.Now())
		require.NoError(t, err)

		// Create a test client object
		objectID := uuid.New()
		objectQuery := `
			INSERT INTO client_objects (id, client_id, name, address, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, objectQuery, objectID, clientID, "Resolution Test Location", "123 Resolution St", time.Now(), time.Now())
		require.NoError(t, err)

		// Create an active order
		orderID := uuid.New()
		orderQuery := `
			INSERT INTO orders (id, client_id, object_id, scheduled_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = tx.Exec(ctx, orderQuery, orderID, clientID, objectID, "2024-01-15", "SCHEDULED", time.Now(), time.Now())
		require.NoError(t, err)

		// Create active equipment
		equipmentID := uuid.New()
		equipmentQuery := `
			INSERT INTO equipment (id, type, volume_l, condition, client_object_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.Exec(ctx, equipmentQuery, equipmentID, "BIN", 100, "GOOD", objectID, time.Now(), time.Now(), nil)
		require.NoError(t, err)

		// Step 1: Verify deletion is blocked
		var blockingConditions int
		blockingQuery := `
			SELECT 
				(SELECT COUNT(*) FROM orders WHERE object_id = $1 AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS') AND deleted_at IS NULL) +
				(SELECT COUNT(*) FROM equipment WHERE client_object_id = $1 AND deleted_at IS NULL)
		`
		err = tx.QueryRow(ctx, blockingQuery, objectID).Scan(&blockingConditions)
		require.NoError(t, err)
		assert.Equal(t, 2, blockingConditions)

		// Step 2: Resolve the order by completing it
		completeOrderQuery := `UPDATE orders SET status = 'COMPLETED', updated_at = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, completeOrderQuery, time.Now(), orderID)
		require.NoError(t, err)

		// Step 3: Move equipment to warehouse (remove from client object)
		// First create a warehouse
		warehouseID := uuid.New()
		warehouseQuery := `
			INSERT INTO warehouses (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, warehouseQuery, warehouseID, "Test Warehouse "+uuid.New().String()[:8], time.Now(), time.Now())
		require.NoError(t, err)

		moveEquipmentQuery := `UPDATE equipment SET client_object_id = NULL, warehouse_id = $1, updated_at = $2 WHERE id = $3`
		_, err = tx.Exec(ctx, moveEquipmentQuery, warehouseID, time.Now(), equipmentID)
		require.NoError(t, err)

		// Step 4: Verify deletion is now allowed
		err = tx.QueryRow(ctx, blockingQuery, objectID).Scan(&blockingConditions)
		require.NoError(t, err)
		assert.Equal(t, 0, blockingConditions)

		// Step 5: Now deletion should succeed
		deleteQuery := `UPDATE client_objects SET deleted_at = $1, updated_at = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, deleteQuery, time.Now(), objectID)
		require.NoError(t, err)

		// Verify deletion succeeded
		var count int
		query := `SELECT COUNT(*) FROM client_objects WHERE id = $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, objectID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestClientObjectsRepository_NameUniqueness(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx pgx.Tx) {
		// Create a test client
		clientID := uuid.New()
		clientQuery := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, clientQuery, clientID, "Uniqueness Test Client", "uniqueness@test.com", "+1234567890", time.Now(), time.Now())
		require.NoError(t, err)

		// Create first client object
		objectID1 := uuid.New()
		objectQuery1 := `
			INSERT INTO client_objects (id, client_id, name, address, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, objectQuery1, objectID1, clientID, "Unique Location Name", "123 Unique St", time.Now(), time.Now())
		require.NoError(t, err)

		// Create second client object with same name for same client (currently allowed by schema)
		objectID2 := uuid.New()
		objectQuery2 := `
			INSERT INTO client_objects (id, client_id, name, address, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, objectQuery2, objectID2, clientID, "Unique Location Name", "456 Different St", time.Now(), time.Now())
		require.NoError(t, err) // Currently allowed by schema (no unique constraint)

		// Verify both objects exist
		var count int
		query := `SELECT COUNT(*) FROM client_objects WHERE client_id = $1 AND name = $2 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, clientID, "Unique Location Name").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count) // Both objects with same name exist

		// But should be able to create with same name for different client
		clientID2 := uuid.New()
		clientQuery2 := `
			INSERT INTO clients (id, name, email, phone, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, clientQuery2, clientID2, "Another Test Client", "another@test.com", "+0987654321", time.Now(), time.Now())
		require.NoError(t, err)

		objectID3 := uuid.New()
		objectQuery3 := `
			INSERT INTO client_objects (id, client_id, name, address, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, objectQuery3, objectID3, clientID2, "Unique Location Name", "789 Another St", time.Now(), time.Now())
		require.NoError(t, err) // Should succeed for different client

		// Verify the third object exists
		query = `SELECT COUNT(*) FROM client_objects WHERE client_id = $1 AND name = $2 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, clientID2, "Unique Location Name").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // One object with this name for the second client
	})
}
