package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTraceIDHandlerAddsActiveTraceID(t *testing.T) {
	provider, exporter := installTestProvider(t)
	defer shutdownTestProvider(t, provider)

	handler := otelhttp.NewHandler(TraceIDHandler(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})), "GET /test")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test", nil))

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
	if got, want := response.Header().Get(TraceIDHeader), spans[0].SpanContext.TraceID().String(); got != want {
		t.Fatalf("%s = %q, want %q", TraceIDHeader, got, want)
	}
}

func TestTraceIDHandlerOmitsHeaderWithoutActiveSpan(t *testing.T) {
	response := httptest.NewRecorder()
	TraceIDHandler(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test", nil))

	if got := response.Header().Get(TraceIDHeader); got != "" {
		t.Fatalf("%s = %q, want empty without an active span", TraceIDHeader, got)
	}
}

func installTestProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return provider, exporter
}

func shutdownTestProvider(t *testing.T, provider *sdktrace.TracerProvider) {
	t.Helper()
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("tracer provider shutdown: %v", err)
	}
}
