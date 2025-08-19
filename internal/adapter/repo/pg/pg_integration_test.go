//go:build integration

package pg

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var TestPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Use the existing test database created by Makefile
	// This database should be running on localhost:5433 as 'eco-van-test-db'
	dsn := "postgres://app:app@localhost:5433/waste_test?sslmode=disable"

	// Wait for database to be ready (Makefile creates it)
	log.Println("Waiting for test database to be ready...")
	var pool *pgxpool.Pool
	var err error

	// Retry connection with exponential backoff
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			// Test the connection
			if err := pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}

		if i < 9 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to test database after retries: %v", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping test database: %v", err)
	}

	log.Println("Successfully connected to test database")
	TestPool = pool

	// Run tests
	code := m.Run()

	// Clean up
	if TestPool != nil {
		TestPool.Close()
	}

	os.Exit(code)
}

// WithTx runs a test function within a database transaction and automatically rolls back
func WithTx(t *testing.T, fn func(ctx context.Context, tx pgx.Tx)) {
	t.Helper()

	ctx := context.Background()
	tx, err := TestPool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	fn(ctx, tx)
}

// MakeClient creates a minimal valid client record
func MakeClient(t *testing.T, ctx context.Context, tx pgx.Tx, name string) uuid.UUID {
	t.Helper()

	if name == "" {
		name = "Acme-" + uuid.NewString()[:8]
	}

	clientID := uuid.New()
	query := `
		INSERT INTO clients (id, name, phone, email)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tx.Exec(ctx, query, clientID, name, "+1234567890", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return clientID
}

// MakeClientObject creates a minimal valid client object record
func MakeClientObject(t *testing.T, ctx context.Context, tx pgx.Tx, clientID uuid.UUID, name string) uuid.UUID {
	t.Helper()

	if name == "" {
		name = "Object-" + uuid.NewString()[:8]
	}

	objectID := uuid.New()
	query := `
		INSERT INTO client_objects (id, client_id, name, address)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tx.Exec(ctx, query, objectID, clientID, name, "123 Test St")
	if err != nil {
		t.Fatalf("Failed to create client object: %v", err)
	}

	return objectID
}

// MakeOrder creates a minimal valid order record
func MakeOrder(t *testing.T, ctx context.Context, tx pgx.Tx, clientID, clientObjectID uuid.UUID) uuid.UUID {
	t.Helper()

	orderID := uuid.New()
	query := `
		INSERT INTO orders (id, client_id, object_id, scheduled_date, status)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(ctx, query, orderID, clientID, clientObjectID, "2025-01-01", "DRAFT")
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	return orderID
}
