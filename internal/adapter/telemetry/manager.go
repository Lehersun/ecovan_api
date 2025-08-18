package telemetry

import (
	"context"
	"fmt"

	"eco-van-api/internal/config"
)

// Manager coordinates all telemetry components
type Manager struct {
	Logger  *Logger
	Tracer  *Tracer
	Metrics *Metrics
	config  *config.Config
}

// NewManager creates a new telemetry manager
func NewManager(cfg *config.Config) (*Manager, error) {
	// Initialize logger
	logger := NewLogger(cfg.Telemetry.LogLevel)

	// Initialize tracer
	tracer := NewTracer("eco-van-api", "1.0.0")
	if cfg.Telemetry.EnableTracing && cfg.Telemetry.OTLPEndpoint != "" {
		if err := tracer.InitTracing(cfg.Telemetry.OTLPEndpoint); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
		logger.Info("Tracing initialized with OTLP endpoint: " + cfg.Telemetry.OTLPEndpoint)
	} else {
		logger.Info("Tracing disabled or no OTLP endpoint configured")
	}

	// Initialize metrics
	metrics := NewMetrics()
	if cfg.Telemetry.EnableMetrics {
		if err := metrics.InitMetrics(); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
		logger.Info("Metrics enabled")
	} else {
		logger.Info("Metrics disabled")
	}

	return &Manager{
		Logger:  logger,
		Tracer:  tracer,
		Metrics: metrics,
		config:  cfg,
	}, nil
}

// Shutdown gracefully shuts down all telemetry components
func (m *Manager) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown tracer
	if err := m.Tracer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("tracer shutdown failed: %w", err))
	}

	// Cleanup metrics
	m.Metrics.Cleanup()

	// Log shutdown completion
	m.Logger.Info("Telemetry manager shutdown completed")

	if len(errs) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %v", errs)
	}

	return nil
}

// IsTracingEnabled returns true if tracing is enabled
func (m *Manager) IsTracingEnabled() bool {
	return m.config.Telemetry.EnableTracing && m.config.Telemetry.OTLPEndpoint != ""
}

// IsMetricsEnabled returns true if metrics are enabled
func (m *Manager) IsMetricsEnabled() bool {
	return m.config.Telemetry.EnableMetrics
}

// TestCleanup cleans up telemetry for testing purposes
func (m *Manager) TestCleanup() {
	m.Metrics.Cleanup()
}
