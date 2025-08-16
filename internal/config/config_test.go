package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnv()

	// Set required environment variables
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test HTTP defaults
	if cfg.HTTP.Addr != ":8080" {
		t.Errorf("Expected HTTP.Addr ':8080', got %s", cfg.HTTP.Addr)
	}
	if cfg.HTTP.ReadTimeout != 15*time.Second {
		t.Errorf("Expected HTTP.ReadTimeout 15s, got %v", cfg.HTTP.ReadTimeout)
	}
	if cfg.HTTP.WriteTimeout != 15*time.Second {
		t.Errorf("Expected HTTP.WriteTimeout 15s, got %v", cfg.HTTP.WriteTimeout)
	}
	if cfg.HTTP.IdleTimeout != 60*time.Second {
		t.Errorf("Expected HTTP.IdleTimeout 60s, got %v", cfg.HTTP.IdleTimeout)
	}
	if cfg.HTTP.MaxBodyBytes != 10*1024*1024 {
		t.Errorf("Expected HTTP.MaxBodyBytes 10MB, got %d", cfg.HTTP.MaxBodyBytes)
	}
	if len(cfg.HTTP.CORSOrigins) != 1 || cfg.HTTP.CORSOrigins[0] != "*" {
		t.Errorf("Expected HTTP.CORSOrigins ['*'], got %v", cfg.HTTP.CORSOrigins)
	}

	// Test DB defaults
	if cfg.DB.MaxConns != 10 {
		t.Errorf("Expected DB.MaxConns 10, got %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 2 {
		t.Errorf("Expected DB.MinConns 2, got %d", cfg.DB.MinConns)
	}
	if cfg.DB.MaxConnLifetime != 30*time.Minute {
		t.Errorf("Expected DB.MaxConnLifetime 30m, got %v", cfg.DB.MaxConnLifetime)
	}
	if cfg.DB.MaxConnIdleTime != 10*time.Minute {
		t.Errorf("Expected DB.MaxConnIdleTime 10m, got %v", cfg.DB.MaxConnIdleTime)
	}

	// Test Auth defaults
	if cfg.Auth.AccessTTL != 15*time.Minute {
		t.Errorf("Expected Auth.AccessTTL 15m, got %v", cfg.Auth.AccessTTL)
	}
	if cfg.Auth.RefreshTTL != 720*time.Hour {
		t.Errorf("Expected Auth.RefreshTTL 720h, got %v", cfg.Auth.RefreshTTL)
	}

	// Test Telemetry defaults
	if cfg.Telemetry.LogLevel != "info" {
		t.Errorf("Expected Telemetry.LogLevel 'info', got %s", cfg.Telemetry.LogLevel)
	}
	if cfg.Telemetry.EnableMetrics != true {
		t.Errorf("Expected Telemetry.EnableMetrics true, got %v", cfg.Telemetry.EnableMetrics)
	}
	if cfg.Telemetry.EnableTracing != false {
		t.Errorf("Expected Telemetry.EnableTracing false, got %v", cfg.Telemetry.EnableTracing)
	}

	// Test Photos defaults
	if cfg.Photos.Dir != "/photos" {
		t.Errorf("Expected Photos.Dir '/photos', got %s", cfg.Photos.Dir)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")

	// Set environment variables to test overrides
	os.Setenv("HTTP_ADDR", ":9090")
	os.Setenv("HTTP_READ_TIMEOUT", "30s")
	os.Setenv("DB_MAX_CONNS", "20")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENABLE_TRACING", "true")
	os.Setenv("PHOTOS_DIR", "/tmp/photos")

	// Clean up after test
	defer func() {
		os.Unsetenv("HTTP_ADDR")
		os.Unsetenv("HTTP_READ_TIMEOUT")
		os.Unsetenv("DB_MAX_CONNS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_TRACING")
		os.Unsetenv("PHOTOS_DIR")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test overrides
	if cfg.HTTP.Addr != ":9090" {
		t.Errorf("Expected HTTP.Addr ':9090', got %s", cfg.HTTP.Addr)
	}
	if cfg.HTTP.ReadTimeout != 30*time.Second {
		t.Errorf("Expected HTTP.ReadTimeout 30s, got %v", cfg.HTTP.ReadTimeout)
	}
	if cfg.DB.MaxConns != 20 {
		t.Errorf("Expected DB.MaxConns 20, got %d", cfg.DB.MaxConns)
	}
	if cfg.Telemetry.LogLevel != "debug" {
		t.Errorf("Expected Telemetry.LogLevel 'debug', got %s", cfg.Telemetry.LogLevel)
	}
	if cfg.Telemetry.EnableTracing != true {
		t.Errorf("Expected Telemetry.EnableTracing true, got %v", cfg.Telemetry.EnableTracing)
	}
	if cfg.Photos.Dir != "/tmp/photos" {
		t.Errorf("Expected Photos.Dir '/tmp/photos', got %s", cfg.Photos.Dir)
	}
}

func TestLoad_RequiredFields(t *testing.T) {
	// Clear environment variables
	clearEnv()

	// Test without required fields
	_, err := Load()
	if err == nil {
		t.Fatal("Expected error when DB_DSN and JWT_SECRET are missing")
	}

	// Test with only DB_DSN
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	defer os.Unsetenv("DB_DSN")

	_, err = Load()
	if err == nil {
		t.Fatal("Expected error when JWT_SECRET is missing")
	}

	// Test with both required fields
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error when both required fields are set, got: %v", err)
	}

	if cfg.DB.DSN != "postgres://test:test@localhost:5432/test" {
		t.Errorf("Expected DB.DSN to be set, got %s", cfg.DB.DSN)
	}
	if cfg.Auth.JWTSecret != "test-secret" {
		t.Errorf("Expected Auth.JWTSecret to be set, got %s", cfg.Auth.JWTSecret)
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	// Test single origin
	os.Setenv("CORS_ORIGINS", "https://example.com")
	defer os.Unsetenv("CORS_ORIGINS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.HTTP.CORSOrigins) != 1 || cfg.HTTP.CORSOrigins[0] != "https://example.com" {
		t.Errorf("Expected CORS_ORIGINS ['https://example.com'], got %v", cfg.HTTP.CORSOrigins)
	}

	// Test multiple origins
	os.Setenv("CORS_ORIGINS", "https://example.com,https://api.example.com")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expected := []string{"https://example.com", "https://api.example.com"}
	if len(cfg.HTTP.CORSOrigins) != 2 {
		t.Errorf("Expected 2 CORS origins, got %d", len(cfg.HTTP.CORSOrigins))
	}
	for i, origin := range expected {
		if cfg.HTTP.CORSOrigins[i] != origin {
			t.Errorf("Expected CORS origin %s at index %d, got %s", origin, i, cfg.HTTP.CORSOrigins[i])
		}
	}
}

func TestLoad_DurationParsing(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_DSN", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	// Test various duration formats
	testCases := []struct {
		envVar   string
		value    string
		expected time.Duration
	}{
		{"HTTP_READ_TIMEOUT", "30s", 30 * time.Second},
		{"HTTP_WRITE_TIMEOUT", "2m", 2 * time.Minute},
		{"HTTP_IDLE_TIMEOUT", "1h", 1 * time.Hour},
		{"DB_MAX_CONN_LIFETIME", "45m", 45 * time.Minute},
		{"ACCESS_TTL", "1h30m", 90 * time.Minute},
	}

	for _, tc := range testCases {
		os.Setenv(tc.envVar, tc.value)
	}

	// Clean up environment variables after test
	defer func() {
		for _, tc := range testCases {
			os.Unsetenv(tc.envVar)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the durations were parsed correctly
	if cfg.HTTP.ReadTimeout != 30*time.Second {
		t.Errorf("Expected HTTP.ReadTimeout 30s, got %v", cfg.HTTP.ReadTimeout)
	}
	if cfg.HTTP.WriteTimeout != 2*time.Minute {
		t.Errorf("Expected HTTP.WriteTimeout 2m, got %v", cfg.HTTP.WriteTimeout)
	}
	if cfg.HTTP.IdleTimeout != 1*time.Hour {
		t.Errorf("Expected HTTP.IdleTimeout 1h, got %v", cfg.HTTP.IdleTimeout)
	}
	if cfg.DB.MaxConnLifetime != 45*time.Minute {
		t.Errorf("Expected DB.MaxConnLifetime 45m, got %v", cfg.DB.MaxConnLifetime)
	}
	if cfg.Auth.AccessTTL != 90*time.Minute {
		t.Errorf("Expected Auth.AccessTTL 1h30m, got %v", cfg.Auth.AccessTTL)
	}
}

// Helper function to clear environment variables
func clearEnv() {
	envVars := []string{
		"HTTP_ADDR", "HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT",
		"HTTP_MAX_BODY", "CORS_ORIGINS", "DB_DSN", "DB_MAX_CONNS", "DB_MIN_CONNS",
		"DB_MAX_CONN_LIFETIME", "DB_MAX_CONN_IDLE", "JWT_SECRET", "ACCESS_TTL",
		"REFRESH_TTL", "LOG_LEVEL", "OTLP_ENDPOINT", "ENABLE_METRICS", "ENABLE_TRACING",
		"PHOTOS_DIR",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
