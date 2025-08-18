package telemetry

import (
	"fmt"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test with valid log level
	logger := NewLogger("debug")
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	// Test with invalid log level (should default to info)
	logger = NewLogger("invalid")
	if logger == nil {
		t.Fatal("Expected logger to be created with invalid level, got nil")
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	logger := NewLogger("info")
	requestID := "test-request-123"

	loggerWithID := logger.WithRequestID(requestID)
	if loggerWithID == nil {
		t.Fatal("Expected logger with request ID to be created, got nil")
	}

	// Verify it's a different instance
	if logger == loggerWithID {
		t.Error("Expected different logger instance")
	}
}

func TestLogger_WithHTTPRequest(t *testing.T) {
	logger := NewLogger("info")
	method := "GET"
	path := "/test"
	status := 200
	duration := 150 * time.Millisecond

	loggerWithHTTP := logger.WithHTTPRequest(method, path, status, duration)
	if loggerWithHTTP == nil {
		t.Fatal("Expected logger with HTTP request to be created, got nil")
	}

	// Verify it's a different instance
	if logger == loggerWithHTTP {
		t.Error("Expected different logger instance")
	}
}

func TestLogger_AccessLog(t *testing.T) {
	logger := NewLogger("info")
	method := "POST"
	path := "/api/v1/test"
	status := 201
	duration := 250 * time.Millisecond
	requestID := "test-request-456"

	// This should not panic
	logger.AccessLog(method, path, status, duration, requestID)
}

func TestLogger_LogLevels(t *testing.T) {
	logger := NewLogger("debug")

	// Test all log levels (should not panic)
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message", nil)
	logger.Error("Error with error", fmt.Errorf("test error"))
}

func TestLogger_GetZerolog(t *testing.T) {
	logger := NewLogger("info")
	zerologLogger := logger.GetZerolog()

	// Just verify the method doesn't panic and returns something
	_ = zerologLogger
}
