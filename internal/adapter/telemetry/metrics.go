package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics wraps Prometheus metrics with additional functionality
type Metrics struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	enabled              bool
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		enabled: false,
	}
}

// InitMetrics initializes Prometheus metrics
func (m *Metrics) InitMetrics() error {
	// HTTP requests total counter
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTP requests in flight gauge
	m.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"method"},
	)

	// Register metrics
	if err := prometheus.Register(m.httpRequestsTotal); err != nil {
		return err
	}
	if err := prometheus.Register(m.httpRequestDuration); err != nil {
		return err
	}
	if err := prometheus.Register(m.httpRequestsInFlight); err != nil {
		return err
	}

	m.enabled = true
	return nil
}

// Cleanup unregisters all metrics (useful for testing)
func (m *Metrics) Cleanup() {
	if m.enabled {
		prometheus.Unregister(m.httpRequestsTotal)
		prometheus.Unregister(m.httpRequestDuration)
		prometheus.Unregister(m.httpRequestsInFlight)
		m.enabled = false
	}
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration float64) {
	if !m.enabled {
		return
	}

	m.httpRequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// IncRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncRequestsInFlight(method string) {
	if !m.enabled {
		return
	}
	m.httpRequestsInFlight.WithLabelValues(method).Inc()
}

// DecRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecRequestsInFlight(method string) {
	if !m.enabled {
		return
	}
	m.httpRequestsInFlight.WithLabelValues(method).Dec()
}

// GetHandler returns the Prometheus HTTP handler
func (m *Metrics) GetHandler() http.Handler {
	return promhttp.Handler()
}

// IsEnabled returns true if metrics are enabled
func (m *Metrics) IsEnabled() bool {
	return m.enabled
}
