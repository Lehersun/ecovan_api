package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	// HTTP timeouts
	DefaultHTTPReadTimeout  = 15 * time.Second
	DefaultHTTPWriteTimeout = 15 * time.Second
	DefaultHTTPIdleTimeout  = 60 * time.Second
	DefaultHTTPMaxBody      = 10 * 1024 * 1024 // 10MB

	// Database connection pool
	DefaultDBMaxConns        = 10
	DefaultDBMinConns        = 2
	DefaultDBMaxConnLifetime = 30 * time.Minute
	DefaultDBMaxConnIdle     = 10 * time.Minute

	// Authentication TTLs
	DefaultAccessTTL  = 15 * time.Minute
	DefaultRefreshTTL = 720 * time.Hour // 30 days
)

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env file doesn't exist
		_ = err // ignore error
	}

	cfg := &Config{
		HTTP:      loadHTTPConfig(),
		DB:        loadDBConfig(),
		Auth:      loadAuthConfig(),
		Telemetry: loadTelemetryConfig(),
		Photos:    loadPhotosConfig(),
	}

	// Validate required fields
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadHTTPConfig loads HTTP configuration with defaults
func loadHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Addr:         getEnv("HTTP_ADDR", ":8080"),
		ReadTimeout:  getEnvAsDuration("HTTP_READ_TIMEOUT", DefaultHTTPReadTimeout),
		WriteTimeout: getEnvAsDuration("HTTP_WRITE_TIMEOUT", DefaultHTTPWriteTimeout),
		IdleTimeout:  getEnvAsDuration("HTTP_IDLE_TIMEOUT", DefaultHTTPIdleTimeout),
		CORSOrigins:  getEnvAsStringSlice("CORS_ORIGINS", []string{"*"}),
		MaxBodyBytes: getEnvAsInt64("HTTP_MAX_BODY", DefaultHTTPMaxBody),
	}
}

// loadDBConfig loads database configuration with defaults
func loadDBConfig() DBConfig {
	return DBConfig{
		DSN:             getEnv("DB_DSN", ""),
		MaxConns:        getEnvAsInt("DB_MAX_CONNS", DefaultDBMaxConns),
		MinConns:        getEnvAsInt("DB_MIN_CONNS", DefaultDBMinConns),
		MaxConnLifetime: getEnvAsDuration("DB_MAX_CONN_LIFETIME", DefaultDBMaxConnLifetime),
		MaxConnIdleTime: getEnvAsDuration("DB_MAX_CONN_IDLE", DefaultDBMaxConnIdle),
	}
}

// loadAuthConfig loads authentication configuration with defaults
func loadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:  getEnv("JWT_SECRET", ""),
		AccessTTL:  getEnvAsDuration("ACCESS_TTL", DefaultAccessTTL),
		RefreshTTL: getEnvAsDuration("REFRESH_TTL", DefaultRefreshTTL),
	}
}

// loadTelemetryConfig loads telemetry configuration with defaults
func loadTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		OTLPEndpoint:  getEnv("OTLP_ENDPOINT", ""),
		EnableMetrics: getEnvAsBool("ENABLE_METRICS", true),
		EnableTracing: getEnvAsBool("ENABLE_TRACING", false),
	}
}

// loadPhotosConfig loads photos configuration with defaults
func loadPhotosConfig() PhotosConfig {
	return PhotosConfig{
		Dir: getEnv("PHOTOS_DIR", "/photos"),
	}
}

// validateConfig validates required configuration fields
func validateConfig(cfg *Config) error {
	if cfg.DB.DSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

// Helper functions for environment variable loading

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		if value == "*" {
			return []string{"*"}
		}
		return strings.Split(value, ",")
	}
	return defaultValue
}
