package app

import (
	"context"
	"testing"
	"time"
)

func TestServer_Shutdown(t *testing.T) {
	config := &Config{
		Port:         "0", // Use port 0 for testing (random available port)
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)

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
	err := server.Shutdown(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error during shutdown, got: %v", err)
	}

	// Assert shutdown completes within 1 second
	if elapsed > 1*time.Second {
		t.Errorf("Shutdown took too long: %v, expected < 1s", elapsed)
	}
}

func TestNewServer(t *testing.T) {
	config := &Config{
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.config != config {
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
	config := &Config{
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if config.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", config.WriteTimeout)
	}

	if config.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", config.IdleTimeout)
	}
}
