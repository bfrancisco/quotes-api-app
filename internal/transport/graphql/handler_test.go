package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	"github.com/bfrancisco/quotes-api-app/internal/telemetry"
	graphqltransport "github.com/bfrancisco/quotes-api-app/internal/transport/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/transport/graphql/generated"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func newServer() http.Handler {
	quoteService := service.NewQuoteService(memory.NewInMemoryRepository())
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: graphqltransport.NewResolver(quoteService),
	}))
	server.SetErrorPresenter(graphqltransport.ErrorPresenter)
	return otelhttp.NewHandler(telemetry.TraceIDHandler(server), "POST /graphql/query")
}

func execute(t *testing.T, server http.Handler, query string) graphQLResponse {
	t.Helper()

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var result graphQLResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return result
}

func assertNoErrors(t *testing.T, response graphQLResponse) {
	t.Helper()
	if len(response.Errors) != 0 {
		t.Fatalf("errors = %+v, want none", response.Errors)
	}
}

func assertErrorCode(t *testing.T, response graphQLResponse, want string) {
	t.Helper()
	if len(response.Errors) != 1 {
		t.Fatalf("error count = %d, want 1", len(response.Errors))
	}
	if code, _ := response.Errors[0].Extensions["code"].(string); code != want {
		t.Fatalf("error code = %q, want %q", code, want)
	}
}

func createQuote(t *testing.T, server http.Handler, text, author string) string {
	t.Helper()

	response := execute(t, server, `mutation { createQuote(input: { text: "`+text+`", author: "`+author+`" }) { id } }`)
	assertNoErrors(t, response)

	var data struct {
		CreateQuote struct {
			ID string `json:"id"`
		} `json:"createQuote"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if data.CreateQuote.ID == "" {
		t.Fatal("created quote ID is empty")
	}
	return data.CreateQuote.ID
}

func TestGraphQLCreatedAtSerialization(t *testing.T) {
	server := newServer()
	id := createQuote(t, server, "First quote", "Author A")

	get := execute(t, server, `{ quote(id: "`+id+`") { createdAt } }`)
	assertNoErrors(t, get)

	var data struct {
		Quote struct {
			CreatedAt string `json:"createdAt"`
		} `json:"quote"`
	}
	if err := json.Unmarshal(get.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if _, err := time.Parse(time.RFC3339Nano, data.Quote.CreatedAt); err != nil {
		t.Fatalf("createdAt = %q, want RFC3339 timestamp: %v", data.Quote.CreatedAt, err)
	}
}

func TestGraphQLRequestContinuesTraceAndReturnsTraceID(t *testing.T) {
	provider, exporter := installGraphQLTestProvider(t)
	defer shutdownGraphQLTestProvider(t, provider)

	request := httptest.NewRequest(http.MethodPost, "/graphql/query", bytes.NewBufferString(`{"query":"{ health { status } }"}`))
	request.Header.Set("Content-Type", "application/json")
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	otel.GetTextMapPropagator().Inject(trace.ContextWithRemoteSpanContext(request.Context(), parent), propagation.HeaderCarrier(request.Header))

	response := httptest.NewRecorder()
	newServer().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
	span := spans[0]
	if span.Name != "POST /graphql/query" {
		t.Fatalf("span name = %q, want stable operation endpoint name", span.Name)
	}
	if span.Parent.TraceID() != parent.TraceID() || span.Parent.SpanID() != parent.SpanID() {
		t.Fatalf("span parent = %v, want propagated remote parent %v", span.Parent, parent)
	}
	if got, want := response.Header().Get(telemetry.TraceIDHeader), span.SpanContext.TraceID().String(); got != want {
		t.Fatalf("%s = %q, want %q", telemetry.TraceIDHeader, got, want)
	}
	if graphQLSpanHasAttributeValue(span.Attributes, `{ health { status } }`) {
		t.Fatalf("span attributes = %v, must not include raw GraphQL query", span.Attributes)
	}
}

func TestGraphQLErrorResponseIncludesTraceID(t *testing.T) {
	provider, exporter := installGraphQLTestProvider(t)
	defer shutdownGraphQLTestProvider(t, provider)

	request := httptest.NewRequest(http.MethodPost, "/graphql/query", bytes.NewBufferString(`{"query":"{ quote(id: \"not-a-uuid\") { id } }"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newServer().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get(telemetry.TraceIDHeader); got == "" {
		t.Fatalf("%s is empty, want active trace ID", telemetry.TraceIDHeader)
	}
	if spans := exporter.GetSpans(); len(spans) != 1 {
		t.Fatalf("exported span count = %d, want 1", len(spans))
	}
}

func installGraphQLTestProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return provider, exporter
}

func shutdownGraphQLTestProvider(t *testing.T, provider *sdktrace.TracerProvider) {
	t.Helper()
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("tracer provider shutdown: %v", err)
	}
}

func graphQLSpanHasAttributeValue(attributes []attribute.KeyValue, value string) bool {
	for _, attribute := range attributes {
		if attribute.Value.AsString() == value {
			return true
		}
	}
	return false
}
