package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eco-van-api/internal/adapter/telemetry"
)

func TestMiddleware_RequestID(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(RequestIDKey)
		if requestID == nil {
			t.Error("Expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Test middleware
	middleware := mw.RequestID()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check response headers
	if w.Header().Get(RequestIDHeader) == "" {
		t.Error("Expected request ID in response headers")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_RequestID_Existing(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(RequestIDKey)
		if requestID != "existing-id" {
			t.Errorf("Expected request ID 'existing-id', got %v", requestID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Test middleware with existing request ID
	middleware := mw.RequestID()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set(RequestIDHeader, "existing-id")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check response headers
	if w.Header().Get(RequestIDHeader) != "existing-id" {
		t.Errorf("Expected request ID 'existing-id' in response headers, got %s", w.Header().Get(RequestIDHeader))
	}
}

func TestMiddleware_Recover(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Test middleware
	middleware := mw.Recover()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	// This should not panic
	middleware.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestMiddleware_AccessLog(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test middleware
	middleware := mw.AccessLog()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_CORS(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test CORS middleware with wildcard origins
	middleware := mw.CORS([]string{"*"})(handler)

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check CORS headers
	expectedOrigin := "https://example.com"
	if w.Header().Get("Access-Control-Allow-Origin") != expectedOrigin {
		t.Errorf("Expected CORS origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test actual request
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", expectedOrigin)
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != expectedOrigin {
		t.Errorf("Expected CORS origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMiddleware_CORS_SpecificOrigins(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test CORS middleware with specific origins
	origins := []string{"https://example.com", "https://api.example.com"}
	middleware := mw.CORS(origins)(handler)

	// Test allowed origin
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	expectedOrigin := "https://example.com"
	if w.Header().Get("Access-Control-Allow-Origin") != expectedOrigin {
		t.Errorf("Expected CORS origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test disallowed origin
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://malicious.com")
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no CORS origin header for disallowed origin, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMiddleware_RateLimit(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test middleware (should not panic)
	middleware := mw.RateLimit()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_MetricsInFlight(t *testing.T) {
	// Create test telemetry components
	logger := telemetry.NewLogger("info")
	tracer := telemetry.NewTracer("test", "1.0.0")
	metrics := telemetry.NewMetrics()

	mw := NewMiddleware(logger, tracer, metrics)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test middleware (should not panic)
	middleware := mw.MetricsInFlight()(handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
