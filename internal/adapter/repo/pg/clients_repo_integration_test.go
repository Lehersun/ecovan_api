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
		// Create multiple test clients
		clients := []struct {
			name  string
			email string
			phone string
		}{
			{"Alpha Company", "alpha@company.com", "+1111111111"},
			{"Beta Corporation", "beta@corp.com", "+2222222222"},
			{"Gamma Industries", "gamma@industries.com", "+3333333333"},
			{"Delta Services", "delta@services.com", "+4444444444"},
		}

		for _, client := range clients {
			query := `
				INSERT INTO clients (id, name, email, phone, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err := tx.Exec(ctx, query, uuid.New(), client.name, client.email, client.phone, time.Now(), time.Now())
			require.NoError(t, err)
		}

		// Test search by name
		query := `SELECT COUNT(*) FROM clients WHERE name ILIKE $1 AND deleted_at IS NULL`
		var count int
		err := tx.QueryRow(ctx, query, "%Company%").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only Alpha Company (Beta Corporation doesn't contain "Company")

		// Test search by email
		query = `SELECT COUNT(*) FROM clients WHERE email ILIKE $1 AND deleted_at IS NULL`
		err = tx.QueryRow(ctx, query, "%corp%").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Beta Corporation

		// Test pagination
		query = `SELECT COUNT(*) FROM clients WHERE deleted_at IS NULL`
		err = tx.QueryRow(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 4, count)

		// Test pagination with limit
		query = `SELECT name FROM clients WHERE deleted_at IS NULL ORDER BY name LIMIT 2`
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
		assert.Equal(t, "Alpha Company", names[0])
		assert.Equal(t, "Beta Corporation", names[1])
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
