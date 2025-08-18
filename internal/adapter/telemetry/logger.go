package telemetry

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger instance with the specified log level
func NewLogger(level string) *Logger {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(logLevel)

	// Create logger with JSON output (more reliable than console writer)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &Logger{
		logger: logger,
	}
}

// WithRequestID adds request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
	}
}

// WithHTTPRequest adds HTTP request metadata to the logger context
func (l *Logger) WithHTTPRequest(method, path string, status int, duration time.Duration) *Logger {
	return &Logger{
		logger: l.logger.With().
			Str("method", method).
			Str("path", path).
			Str("status", fmt.Sprintf("%d", status)).
			Int64("dur_ms", duration.Milliseconds()).
			Logger(),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string, err error) {
	event := l.logger.Panic()
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

// AccessLog logs HTTP access with structured fields
func (l *Logger) AccessLog(method, path string, status int, duration time.Duration, requestID string) {
	l.WithHTTPRequest(method, path, status, duration).
		WithRequestID(requestID).
		Info("HTTP request completed")
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}
