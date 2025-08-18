package pg

import (
	"context"
	"fmt"

	"eco-van-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	mock bool
}

// NewDB creates a new database connection pool
func NewDB(cfg *config.Config) (*DB, error) {
	// Parse the connection string
	config, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = int32(cfg.DB.MaxConns)
	config.MinConns = int32(cfg.DB.MinConns)
	config.MaxConnLifetime = cfg.DB.MaxConnLifetime
	config.MaxConnIdleTime = cfg.DB.MaxConnIdleTime

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		pool: pool,
		cfg:  cfg,
		mock: false,
	}, nil
}

// NewMockDB creates a mock database for testing
func NewMockDB() *DB {
	return &DB{
		pool: nil,
		cfg:  nil,
		mock: true,
	}
}

// GetPool returns the underlying pgxpool.Pool
func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Ping checks if the database is accessible
func (db *DB) Ping(ctx context.Context) error {
	if db.mock {
		return nil // Mock always succeeds
	}
	return db.pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// IsHealthy returns true if the database is healthy
func (db *DB) IsHealthy(ctx context.Context) bool {
	return db.Ping(ctx) == nil
}

// GetStats returns connection pool statistics
func (db *DB) GetStats() *pgxpool.Stat {
	if db.mock {
		return &pgxpool.Stat{}
	}
	return db.pool.Stat()
}
