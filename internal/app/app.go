package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eco-van-api/internal/config"
)

const (
	shutdownTimeout = 5 * time.Second
)

// App represents the main application
type App struct {
	server *Server
	config *config.Config
}

// New creates a new App instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	server := NewServer(cfg)

	return &App{
		server: server,
		config: cfg,
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

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
