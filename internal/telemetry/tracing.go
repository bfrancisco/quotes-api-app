// Package telemetry configures process-wide OpenTelemetry tracing.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config identifies the API process in every emitted span.
type Config struct {
	ServiceName           string
	ServiceVersion        string
	DeploymentEnvironment string
}

// ShutdownFunc flushes queued spans and closes the telemetry exporter.
type ShutdownFunc func(context.Context) error

// Setup configures global tracing. The OTLP/HTTP exporter reads its endpoint URL, authentication,
// TLS, and timeout from the standard OTEL_EXPORTER_OTLP_* environment variables.
func Setup(ctx context.Context, config Config) (ShutdownFunc, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	provider, err := NewTracerProvider(config, exporter)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, err
	}

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return provider.Shutdown, nil
}

// NewTracerProvider creates a batching trace provider without changing global OpenTelemetry
// state. Tests use this function with an in-memory exporter rather than a live Collector.
func NewTracerProvider(config Config, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	if config.ServiceName == "" {
		return nil, fmt.Errorf("telemetry service name must not be empty")
	}
	if exporter == nil {
		return nil, fmt.Errorf("trace exporter must not be nil")
	}

	res, err := newResource(config)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		// Full sampling is implemented since we expect very low traffic.
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithBatcher(exporter),
	), nil
}

func newResource(config Config) (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		attribute.String("service.name", config.ServiceName),
	}
	if config.ServiceVersion != "" {
		attributes = append(attributes, attribute.String("service.version", config.ServiceVersion))
	}
	if config.DeploymentEnvironment != "" {
		attributes = append(attributes, attribute.String("deployment.environment.name", config.DeploymentEnvironment))
	}

	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(attributes...))
	if err != nil {
		return nil, fmt.Errorf("build telemetry resource: %w", err)
	}
	return res, nil
}
