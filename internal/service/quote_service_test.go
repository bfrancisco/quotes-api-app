package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const validQuoteID = "550e8400-e29b-41d4-a716-446655440000"

type fakeQuoteRepository struct {
	createQuote       func(context.Context, model.QuoteCreateInput) (model.Quote, error)
	listQuotes        func(context.Context) ([]model.Quote, error)
	getQuoteByID      func(context.Context, string) (model.Quote, error)
	getQuotesByAuthor func(context.Context, string) ([]model.Quote, error)
	getRandomQuote    func(context.Context) (model.Quote, error)
	updateQuote       func(context.Context, model.QuoteUpdateInput) (model.Quote, error)
	deleteQuote       func(context.Context, string) error
}

func (f fakeQuoteRepository) CreateQuote(ctx context.Context, input model.QuoteCreateInput) (model.Quote, error) {
	return f.createQuote(ctx, input)
}

func (f fakeQuoteRepository) ListQuotes(ctx context.Context) ([]model.Quote, error) {
	return f.listQuotes(ctx)
}

func (f fakeQuoteRepository) GetQuoteByID(ctx context.Context, id string) (model.Quote, error) {
	return f.getQuoteByID(ctx, id)
}

func (f fakeQuoteRepository) GetQuotesByAuthor(ctx context.Context, author string) ([]model.Quote, error) {
	return f.getQuotesByAuthor(ctx, author)
}

func (f fakeQuoteRepository) GetRandomQuote(ctx context.Context) (model.Quote, error) {
	return f.getRandomQuote(ctx)
}

func (f fakeQuoteRepository) UpdateQuote(ctx context.Context, input model.QuoteUpdateInput) (model.Quote, error) {
	return f.updateQuote(ctx, input)
}

func (f fakeQuoteRepository) DeleteQuote(ctx context.Context, id string) error {
	return f.deleteQuote(ctx, id)
}

func newFakeRepository() fakeQuoteRepository {
	return fakeQuoteRepository{
		createQuote:       func(context.Context, model.QuoteCreateInput) (model.Quote, error) { return model.Quote{}, nil },
		listQuotes:        func(context.Context) ([]model.Quote, error) { return nil, nil },
		getQuoteByID:      func(context.Context, string) (model.Quote, error) { return model.Quote{}, nil },
		getQuotesByAuthor: func(context.Context, string) ([]model.Quote, error) { return nil, nil },
		getRandomQuote:    func(context.Context) (model.Quote, error) { return model.Quote{}, nil },
		updateQuote:       func(context.Context, model.QuoteUpdateInput) (model.Quote, error) { return model.Quote{}, nil },
		deleteQuote:       func(context.Context, string) error { return nil },
	}
}

func testQuote(id string) model.Quote {
	return model.Quote{ID: id, Text: "Test quote", Author: "Test author", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func TestCreateQuoteValidatesBeforeCallingRepository(t *testing.T) {
	repositoryCalled := false
	repository := newFakeRepository()
	repository.createQuote = func(context.Context, model.QuoteCreateInput) (model.Quote, error) {
		repositoryCalled = true
		return model.Quote{}, nil
	}

	_, err := service.NewQuoteService(repository).CreateQuote(context.Background(), model.QuoteCreateInput{Text: " ", Author: "Author"})
	if !errors.Is(err, model.ErrInvalidQuoteText) {
		t.Fatalf("CreateQuote() error = %v, want %v", err, model.ErrInvalidQuoteText)
	}
	if repositoryCalled {
		t.Fatal("CreateQuote() called repository for invalid input")
	}
}

func TestListQuotesFiltersAndPaginates(t *testing.T) {
	quotes := []model.Quote{testQuote("550e8400-e29b-41d4-a716-446655440001"), testQuote("550e8400-e29b-41d4-a716-446655440002"), testQuote("550e8400-e29b-41d4-a716-446655440003")}
	repository := newFakeRepository()
	repository.getQuotesByAuthor = func(_ context.Context, author string) ([]model.Quote, error) {
		if author != "Test author" {
			t.Fatalf("GetQuotesByAuthor() author = %q, want %q", author, "Test author")
		}
		return quotes, nil
	}

	result, err := service.NewQuoteService(repository).ListQuotes(context.Background(), service.QuoteListInput{Author: "Test author", Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}
	if result.Count != 1 || result.Limit != 1 || result.Offset != 1 || result.Quotes[0].ID != quotes[1].ID {
		t.Fatalf("ListQuotes() result = %#v, want second quote with limit 1 and offset 1", result)
	}
}

func TestListQuotesUsesDefaultLimitAndRejectsInvalidPaging(t *testing.T) {
	repository := newFakeRepository()
	repository.listQuotes = func(context.Context) ([]model.Quote, error) {
		return []model.Quote{testQuote(validQuoteID)}, nil
	}
	quoteService := service.NewQuoteService(repository)

	result, err := quoteService.ListQuotes(context.Background(), service.QuoteListInput{})
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}
	if result.Limit != 20 || result.Count != 1 {
		t.Fatalf("ListQuotes() result = %#v, want default limit 20 and count 1", result)
	}

	_, err = quoteService.ListQuotes(context.Background(), service.QuoteListInput{Limit: 101})
	if !errors.Is(err, model.ErrInvalidQuoteListOptions) {
		t.Fatalf("ListQuotes() error = %v, want %v", err, model.ErrInvalidQuoteListOptions)
	}
}

func TestGetAndDeleteQuoteValidateIDBeforeCallingRepository(t *testing.T) {
	getCalled := false
	deleteCalled := false
	repository := newFakeRepository()
	repository.getQuoteByID = func(context.Context, string) (model.Quote, error) {
		getCalled = true
		return model.Quote{}, nil
	}
	repository.deleteQuote = func(context.Context, string) error {
		deleteCalled = true
		return nil
	}
	quoteService := service.NewQuoteService(repository)

	_, getErr := quoteService.GetQuoteByID(context.Background(), "not-a-uuid")
	deleteErr := quoteService.DeleteQuote(context.Background(), "not-a-uuid")
	if !errors.Is(getErr, model.ErrInvalidQuoteID) || !errors.Is(deleteErr, model.ErrInvalidQuoteID) {
		t.Fatalf("invalid ID errors = (%v, %v), want %v", getErr, deleteErr, model.ErrInvalidQuoteID)
	}
	if getCalled || deleteCalled {
		t.Fatal("service called repository for an invalid ID")
	}
}

func TestCreateQuoteCreatesChildSpanAndPropagatesItsContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	defer provider.Shutdown(context.Background())

	repository := newFakeRepository()
	var repositorySpan trace.SpanContext
	repository.createQuote = func(ctx context.Context, _ model.QuoteCreateInput) (model.Quote, error) {
		repositorySpan = trace.SpanContextFromContext(ctx)
		return testQuote(validQuoteID), nil
	}

	parentContext, parentSpan := provider.Tracer("test").Start(context.Background(), "parent")
	_, err := service.NewQuoteService(repository).CreateQuote(parentContext, model.QuoteCreateInput{Text: "Trace me", Author: "Author"})
	parentSpan.End()
	if err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("exported span count = %d, want parent and service spans", len(spans))
	}
	serviceSpan := spans[0]
	if serviceSpan.Name != "quote.create" || serviceSpan.Parent.SpanID() != parentSpan.SpanContext().SpanID() {
		t.Fatalf("service span = %#v, want quote.create child of parent", serviceSpan)
	}
	if repositorySpan.SpanID() != serviceSpan.SpanContext.SpanID() {
		t.Fatalf("repository context span = %s, want service span %s", repositorySpan.SpanID(), serviceSpan.SpanContext.SpanID())
	}
	if serviceSpanHasAttribute(serviceSpan.Attributes, "quote text", "Trace me") || serviceSpanHasAttribute(serviceSpan.Attributes, "quote author", "Author") {
		t.Fatalf("service span attributes = %v, must not include quote content", serviceSpan.Attributes)
	}
}

func TestInvalidQuoteIDIsRecordedAsExpectedDomainOutcome(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	defer provider.Shutdown(context.Background())

	_, err := service.NewQuoteService(newFakeRepository()).GetQuoteByID(context.Background(), "not-a-uuid")
	if !errors.Is(err, model.ErrInvalidQuoteID) {
		t.Fatalf("GetQuoteByID() error = %v, want %v", err, model.ErrInvalidQuoteID)
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 || spans[0].Status.Code != 0 {
		t.Fatalf("span status = %#v, want expected outcome without error status", spans)
	}
}

func serviceSpanHasAttribute(attributes []attribute.KeyValue, key, value string) bool {
	for _, attribute := range attributes {
		if string(attribute.Key) == key && attribute.Value.AsString() == value {
			return true
		}
	}
	return false
}
