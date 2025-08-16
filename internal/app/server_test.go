package app

import (
	"context"
	"os"
	"testing"
	"time"

	"eco-van-api/internal/config"
)

func TestServer_Shutdown(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server := NewServer(cfg)

	// Start server in background
	go func() {
		_ = server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test shutdown with context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	shutdownErr := server.Shutdown(ctx)
	elapsed := time.Since(start)

	if shutdownErr != nil {
		t.Errorf("Expected no error during shutdown, got: %v", shutdownErr)
	}

	// Assert shutdown completes within 1 second
	if elapsed > 1*time.Second {
		t.Errorf("Shutdown took too long: %v, expected < 1s", elapsed)
	}
}

func TestNewServer(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server := NewServer(cfg)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.config != cfg {
		t.Error("Expected server config to match input config")
	}

	if server.router == nil {
		t.Error("Expected router to be initialized")
	}

	if server.server == nil {
		t.Error("Expected http.Server to be initialized")
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.HTTP.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", cfg.HTTP.ReadTimeout)
	}

	if cfg.HTTP.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", cfg.HTTP.WriteTimeout)
	}

	if cfg.HTTP.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", cfg.HTTP.IdleTimeout)
	}
}
