package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	"github.com/bfrancisco/quotes-api-app/internal/telemetry"
	resttransport "github.com/bfrancisco/quotes-api-app/internal/transport/rest"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type quoteEnvelope struct {
	Data struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		Author string `json:"author"`
	} `json:"data"`
}

type quoteListEnvelope struct {
	Data []struct {
		ID     string `json:"id"`
		Author string `json:"author"`
	} `json:"data"`
	Meta struct {
		Count  int `json:"count"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"meta"`
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(
		otelgin.Middleware("quotes-rest-api", otelgin.WithFilter(func(request *http.Request) bool {
			return request.URL.Path != "/v1/health"
		})),
		telemetry.GinTraceIDMiddleware(),
	)
	quoteService := service.NewQuoteService(memory.NewInMemoryRepository())
	resttransport.NewHandler(quoteService).RegisterRoutes(router.Group("/v1"))
	return router
}

func serve(router *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestInvalidCreateJSONResponse(t *testing.T) {
	response := serve(newRouter(), http.MethodPost, "/v1/quotes", `{"text":`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}

	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if envelope.Error.Code != "INVALID_REQUEST_BODY" {
		t.Fatalf("error code = %q, want %q", envelope.Error.Code, "INVALID_REQUEST_BODY")
	}
}

func TestRESTRequestContinuesTraceAndReturnsTraceID(t *testing.T) {
	provider, exporter := installRESTTestProvider(t)
	defer shutdownRESTTestProvider(t, provider)

	request := httptest.NewRequest(http.MethodPost, "/v1/quotes", bytes.NewBufferString(`{"text":"Trace me","author":"Ada"}`))
	request.Header.Set("Content-Type", "application/json")
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	otel.GetTextMapPropagator().Inject(trace.ContextWithRemoteSpanContext(request.Context(), parent), propagation.HeaderCarrier(request.Header))

	response := httptest.NewRecorder()
	newRouter().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
	span := spans[0]
	if span.Name != "POST /v1/quotes" {
		t.Fatalf("span name = %q, want route-oriented name", span.Name)
	}
	if span.Parent.TraceID() != parent.TraceID() || span.Parent.SpanID() != parent.SpanID() {
		t.Fatalf("span parent = %v, want propagated remote parent %v", span.Parent, parent)
	}
	if got, want := response.Header().Get(telemetry.TraceIDHeader), span.SpanContext.TraceID().String(); got != want {
		t.Fatalf("%s = %q, want %q", telemetry.TraceIDHeader, got, want)
	}
}

func TestRESTHealthCheckIsNotTraced(t *testing.T) {
	provider, exporter := installRESTTestProvider(t)
	defer shutdownRESTTestProvider(t, provider)

	response := serve(newRouter(), http.MethodGet, "/v1/health", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if spans := exporter.GetSpans(); len(spans) != 0 {
		t.Fatalf("exported span count = %d, want 0 for health check", len(spans))
	}
	if got := response.Header().Get(telemetry.TraceIDHeader); got != "" {
		t.Fatalf("%s = %q, want empty for untraced health check", telemetry.TraceIDHeader, got)
	}
}

func TestRESTResponseIncludesTraceIDWithoutRequestBodyAttributes(t *testing.T) {
	provider, exporter := installRESTTestProvider(t)
	defer shutdownRESTTestProvider(t, provider)

	response := serve(newRouter(), http.MethodPost, "/v1/quotes", `{"text":"private quote text","author":"private author"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
	if got := response.Header().Get(telemetry.TraceIDHeader); got == "" {
		t.Fatalf("%s is empty, want active trace ID", telemetry.TraceIDHeader)
	}
	if spanHasAttributeValue(spans[0].Attributes, "private quote text") || spanHasAttributeValue(spans[0].Attributes, "private author") {
		t.Fatalf("span attributes = %v, must not include request body values", spans[0].Attributes)
	}
}

func installRESTTestProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return provider, exporter
}

func shutdownRESTTestProvider(t *testing.T, provider *sdktrace.TracerProvider) {
	t.Helper()
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("tracer provider shutdown: %v", err)
	}
}

func spanHasAttributeValue(attributes []attribute.KeyValue, value string) bool {
	for _, attribute := range attributes {
		if attribute.Value.AsString() == value {
			return true
		}
	}
	return false
}
