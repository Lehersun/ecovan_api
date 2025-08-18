package app

import (
	"context"
	"fmt"
	"log"
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
	if err := userRepo.SeedAdminUser(context.Background()); err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	}

	// Run migrations using migrate command
	// For now, we'll skip migrations in the Go code and rely on the migrate command
	// TODO: Implement proper migration handling

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
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, starting graceful shutdown", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, starting graceful shutdown")
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Shutdown telemetry first
	if err := app.telemetry.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during telemetry shutdown: %v", err)
	}

	// Close database
	app.db.Close()

	// Shutdown server
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
