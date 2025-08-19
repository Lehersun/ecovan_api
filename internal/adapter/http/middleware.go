package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"eco-van-api/internal/adapter/telemetry"
)

const (
	RequestIDHeader = "X-Request-ID"
)

type requestIDKey struct{}

var RequestIDKey = requestIDKey{}

// Middleware represents HTTP middleware functions
type Middleware struct {
	logger  *telemetry.Logger
	tracer  *telemetry.Tracer
	metrics *telemetry.Metrics
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(logger *telemetry.Logger, tracer *telemetry.Tracer, metrics *telemetry.Metrics) *Middleware {
	return &Middleware{
		logger:  logger,
		tracer:  tracer,
		metrics: metrics,
	}
}

// RequestID adds a unique request ID to each request
func (m *Middleware) RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Add request ID to request context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Recover recovers from panics and logs the error
func (m *Middleware) Recover() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := r.Context().Value(RequestIDKey)
					if requestID == nil {
						requestID = "unknown"
					}

					// Log the panic
					m.logger.WithRequestID(requestID.(string)).
						Error("HTTP handler panic", fmt.Errorf("panic: %v", err))

					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// AccessLog logs HTTP access with structured fields
func (m *Middleware) AccessLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Get request ID from context
			requestID := r.Context().Value(RequestIDKey)
			if requestID == nil {
				requestID = "unknown"
			}

			// Log access
			m.logger.AccessLog(r.Method, r.URL.Path, wrapped.statusCode, duration, requestID.(string))

			// Record metrics if enabled
			if m.metrics.IsEnabled() {
				m.metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration.Seconds())
			}
		})
	}
}

// Trace adds OpenTelemetry tracing to requests
func (m *Middleware) Trace() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.tracer.IsEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			// Extract trace context from headers
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagationHeaderCarrier(r.Header))

			// Start span
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := m.tracer.StartSpan(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Add trace context to request
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(w, r)

			// Add response status to span
			if wrapped, ok := w.(*responseWriter); ok {
				span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))
			}
		})
	}
}

// CORS adds CORS headers based on configuration
func (m *Middleware) CORS(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle preflight request
			if r.Method == http.MethodOptions {
				origin := r.Header.Get("Origin")
				if m.isAllowedOrigin(origin, origins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusOK)
				return
			}

			// Handle actual request
			origin := r.Header.Get("Origin")
			if m.isAllowedOrigin(origin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit is a placeholder for rate limiting (no logic yet)
func (m *Middleware) RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limiting logic placeholder
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsInFlight tracks in-flight requests for metrics
func (m *Middleware) MetricsInFlight() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m.metrics.IsEnabled() {
				m.metrics.IncRequestsInFlight(r.Method)
				defer m.metrics.DecRequestsInFlight(r.Method)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isAllowedOrigin checks if the origin is allowed
func (m *Middleware) isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	// Allow all origins if "*" is specified
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return true
	}

	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
	}

	return false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// propagationHeaderCarrier implements otel.TextMapCarrier for HTTP headers
type propagationHeaderCarrier http.Header

func (c propagationHeaderCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

func (c propagationHeaderCarrier) Set(key, value string) {
	http.Header(c).Set(key, value)
}

func (c propagationHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
