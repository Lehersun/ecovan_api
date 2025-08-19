package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eco-van-api/internal/adapter/repo/pg"
	"eco-van-api/internal/adapter/telemetry"
	"eco-van-api/internal/config"
)

const (
	shutdownTimeout = 5 * time.Second
)

// App represents the main application
type App struct {
	server    *Server
	config    *config.Config
	telemetry *telemetry.Manager
	db        *pg.DB
}

// New creates a new App instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Initialize telemetry
	telemetry, err := telemetry.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize database
	db, err := pg.NewDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Seed admin user
	userRepo := pg.NewUserRepository(db)
	_ = userRepo.SeedAdminUser(context.Background())

	// Run migrations using migrate command
	// For now, we'll skip migrations in the Go code and rely on the migrate command

	server := NewServer(cfg, telemetry, db)

	return &App{
		server:    server,
		config:    cfg,
		telemetry: telemetry,
		db:        db,
	}, nil
}

// Run starts the application and handles graceful shutdown
func Run(ctx context.Context) error {
	app, err := New()
	if err != nil {
		return err
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start server in goroutine
	go func() {
		if err := app.server.Start(); err != nil {
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		// Received signal, starting graceful shutdown
	case <-ctx.Done():
		// Context cancelled, starting graceful shutdown
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Shutdown telemetry first
	_ = app.telemetry.Shutdown(shutdownCtx)

	// Close database
	app.db.Close()

	// Shutdown server
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	// Server shutdown completed
	return nil
}
