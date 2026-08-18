package telemetry

import (
	"context"
	"errors"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorDetails describes a domain error without exposing its internal cause.
type ErrorDetails struct {
	Code     string
	Message  string
	Expected bool
}

// ClassifyError translates application errors into stable, client-safe error details.
func ClassifyError(err error) ErrorDetails {
	switch {
	case errors.Is(err, model.ErrInvalidQuoteListOptions):
		return ErrorDetails{Code: "INVALID_QUERY_PARAMS", Message: "Invalid query parameters", Expected: true}
	case errors.Is(err, model.ErrInvalidQuoteID):
		return ErrorDetails{Code: "INVALID_QUOTE_ID", Message: "Invalid quote ID", Expected: true}
	case errors.Is(err, model.ErrInvalidQuoteText):
		return ErrorDetails{Code: "INVALID_QUOTE_TEXT", Message: "Invalid quote text", Expected: true}
	case errors.Is(err, model.ErrInvalidQuoteAuthor):
		return ErrorDetails{Code: "INVALID_QUOTE_AUTHOR", Message: "Invalid quote author", Expected: true}
	case errors.Is(err, model.ErrNoFieldsToUpdate):
		return ErrorDetails{Code: "NO_FIELDS_TO_UPDATE", Message: "No fields to update", Expected: true}
	case errors.Is(err, model.ErrQuoteAlreadyExists):
		return ErrorDetails{Code: "QUOTE_ALREADY_EXISTS", Message: "Quote already exists", Expected: true}
	case errors.Is(err, model.ErrQuoteNotFound):
		return ErrorDetails{Code: "QUOTE_NOT_FOUND", Message: "Quote not found", Expected: true}
	default:
		return ErrorDetails{Code: "INTERNAL_ERROR", Message: "Unexpected error"}
	}
}

// RecordError annotates the active span without recording user data or raw internal errors.
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	details := ClassifyError(err)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("error.type", details.Code))
	if details.Expected {
		span.AddEvent("domain.error", trace.WithAttributes(attribute.String("error.type", details.Code)))
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, details.Code)
}
