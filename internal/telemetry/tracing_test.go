package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewTracerProviderExportsSpansWithResourceMetadata(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider, err := NewTracerProvider(Config{
		ServiceName:           "quotes-rest-api",
		ServiceVersion:        "test-version",
		DeploymentEnvironment: "test",
	}, exporter)
	if err != nil {
		t.Fatalf("NewTracerProvider() error = %v, want nil", err)
	}

	ctx, span := provider.Tracer("telemetry-test").Start(context.Background(), "quote.create")
	span.End()
	if err := provider.ForceFlush(ctx); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
	attributes := spans[0].Resource.Attributes()
	if !hasStringAttribute(attributes, "service.name", "quotes-rest-api") ||
		!hasStringAttribute(attributes, "service.version", "test-version") ||
		!hasStringAttribute(attributes, "deployment.environment.name", "test") {
		t.Fatalf("resource attributes = %v, want configured service metadata", attributes)
	}
	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v, want nil", err)
	}
}

func TestNewTracerProviderRejectsMissingServiceName(t *testing.T) {
	if _, err := NewTracerProvider(Config{}, tracetest.NewInMemoryExporter()); err == nil {
		t.Fatal("NewTracerProvider() error = nil, want service name validation error")
	}
}

func hasStringAttribute(attributes []attribute.KeyValue, key, value string) bool {
	for _, attribute := range attributes {
		if string(attribute.Key) == key && attribute.Value.AsString() == value {
			return true
		}
	}
	return false
}
