package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestClassifyErrorReturnsSafeExpectedDomainDetails(t *testing.T) {
	details := ClassifyError(model.ErrQuoteNotFound)
	if details.Code != "QUOTE_NOT_FOUND" || details.Message != "Quote not found" || !details.Expected {
		t.Fatalf("ClassifyError() = %#v, want expected quote-not-found details", details)
	}
}

func TestRecordErrorMarksUnexpectedErrorsOnly(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	defer provider.Shutdown(context.Background())

	ctx, expectedSpan := provider.Tracer("test").Start(context.Background(), "expected")
	RecordError(ctx, model.ErrInvalidQuoteID)
	expectedSpan.End()
	ctx, unexpectedSpan := provider.Tracer("test").Start(context.Background(), "unexpected")
	RecordError(ctx, errors.New("storage credentials expired"))
	unexpectedSpan.End()

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("exported span count = %d, want 2", len(spans))
	}
	if spans[0].Status.Code != codes.Unset {
		t.Fatalf("expected domain status = %v, want unset", spans[0].Status.Code)
	}
	if spans[1].Status.Code != codes.Error || spans[1].Status.Description != "INTERNAL_ERROR" {
		t.Fatalf("unexpected error status = %#v, want INTERNAL_ERROR", spans[1].Status)
	}
}
