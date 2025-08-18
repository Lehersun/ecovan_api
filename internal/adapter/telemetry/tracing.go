package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer with additional functionality
type Tracer struct {
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
}

// NewTracer creates a new tracer instance
func NewTracer(serviceName, serviceVersion string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(serviceName),
	}
}

// InitTracing initializes OpenTelemetry tracing with OTLP exporter
func (t *Tracer) InitTracing(otlpEndpoint string) error {
	if otlpEndpoint == "" {
		// No tracing endpoint provided, use no-op tracer
		return nil
	}

	// Create OTLP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("eco-van-api"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Update tracer
	t.tracer = tp.Tracer("eco-van-api")
	t.tp = tp

	return nil
}

const (
	shutdownTimeout = 5 * time.Second
)

// Shutdown gracefully shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.tp != nil {
		// Give traces time to be exported
		ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		return t.tp.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// GetTracer returns the underlying OpenTelemetry tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// IsEnabled returns true if tracing is enabled
func (t *Tracer) IsEnabled() bool {
	return t.tp != nil
}
