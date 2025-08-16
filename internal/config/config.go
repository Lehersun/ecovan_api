package config

import "time"

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	CORSOrigins  []string
	MaxBodyBytes int64
}

// DBConfig holds database configuration
type DBConfig struct {
	DSN             string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	LogLevel      string
	OTLPEndpoint  string
	EnableMetrics bool
	EnableTracing bool
}

// PhotosConfig holds photo storage configuration
type PhotosConfig struct {
	Dir string
}

// Config holds all application configuration
type Config struct {
	HTTP      HTTPConfig
	DB        DBConfig
	Auth      AuthConfig
	Telemetry TelemetryConfig
	Photos    PhotosConfig
}
