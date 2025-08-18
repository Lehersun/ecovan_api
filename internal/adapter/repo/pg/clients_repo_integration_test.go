//go:build integration

package pg

import (
	"context"
	"testing"

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
		query := `SELECT client_id, client_object_id, status FROM orders WHERE id = $1`
		err := tx.QueryRow(ctx, query, orderID).Scan(&orderClientID, &orderObjectID, &status)
		require.NoError(t, err)

		assert.Equal(t, clientID, orderClientID)
		assert.Equal(t, objectID, orderObjectID)
		assert.Equal(t, "pending", status)
	})
}
