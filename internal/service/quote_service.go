package service

import (
	"context"

	"github.com/bfrancisco/quotes-api-app/internal/helpers"
	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/bfrancisco/quotes-api-app/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const defaultQuoteListLimit = 20 // Indicated in openapi.yaml.

// QuoteListInput specifies optional filtering and pagination for quote listings.
type QuoteListInput struct {
	Author string
	Limit  int
	Offset int
}

// QuoteListResult contains a page of quotes and the effective paging values.
type QuoteListResult struct {
	Quotes []model.Quote
	Count  int
	Limit  int
	Offset int
}

// QuoteService coordinates quote use cases independently of the transport and storage technology.
type QuoteService struct {
	repository repository.QuoteRepository
}

// NewQuoteService creates a service backed by the supplied persistence port.
func NewQuoteService(repository repository.QuoteRepository) *QuoteService {
	return &QuoteService{repository: repository}
}

func (s *QuoteService) CreateQuote(ctx context.Context, input model.QuoteCreateInput) (quote model.Quote, err error) {
	ctx, span := startOperation(ctx, "quote.create")
	defer finishOperation(ctx, span, &err)

	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	return s.repository.CreateQuote(ctx, input)
}

func (s *QuoteService) ListQuotes(ctx context.Context, input QuoteListInput) (result QuoteListResult, err error) {
	ctx, span := startOperation(ctx, "quote.list")
	defer finishOperation(ctx, span, &err)

	limit := input.Limit
	if limit == 0 {
		limit = defaultQuoteListLimit
	}
	span.SetAttributes(
		attribute.Int("quote.list.limit", limit),
		attribute.Int("quote.list.offset", input.Offset),
		attribute.Bool("quote.list.has_author_filter", input.Author != ""),
	)
	if input.Offset < 0 || limit < 1 || limit > 100 {
		return QuoteListResult{}, model.ErrInvalidQuoteListOptions
	}

	var (
		quotes []model.Quote
	)

	if input.Author != "" {
		quotes, err = s.repository.GetQuotesByAuthor(ctx, input.Author)
	} else {
		quotes, err = s.repository.ListQuotes(ctx)
	}
	if err != nil {
		return QuoteListResult{}, err
	}

	start := input.Offset
	if start > len(quotes) {
		start = len(quotes)
	}

	end := start + limit
	if end > len(quotes) {
		end = len(quotes)
	}

	page := quotes[start:end]
	return QuoteListResult{
		Quotes: page,
		Count:  len(page),
		Limit:  limit,
		Offset: input.Offset,
	}, nil
}

func (s *QuoteService) GetQuoteByID(ctx context.Context, id string) (quote model.Quote, err error) {
	ctx, span := startOperation(ctx, "quote.get")
	defer finishOperation(ctx, span, &err)

	if !helpers.IsValidUUID(id) {
		return model.Quote{}, model.ErrInvalidQuoteID
	}

	return s.repository.GetQuoteByID(ctx, id)
}

func (s *QuoteService) GetRandomQuote(ctx context.Context) (quote model.Quote, err error) {
	ctx, span := startOperation(ctx, "quote.random")
	defer finishOperation(ctx, span, &err)

	return s.repository.GetRandomQuote(ctx)
}

func (s *QuoteService) UpdateQuote(ctx context.Context, input model.QuoteUpdateInput) (quote model.Quote, err error) {
	ctx, span := startOperation(ctx, "quote.update")
	defer finishOperation(ctx, span, &err)

	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	return s.repository.UpdateQuote(ctx, input)
}

func (s *QuoteService) DeleteQuote(ctx context.Context, id string) (err error) {
	ctx, span := startOperation(ctx, "quote.delete")
	defer finishOperation(ctx, span, &err)

	if !helpers.IsValidUUID(id) {
		return model.ErrInvalidQuoteID
	}

	return s.repository.DeleteQuote(ctx, id)
}

func startOperation(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("quotes-api/service").Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
}

func finishOperation(ctx context.Context, span trace.Span, err *error) {
	telemetry.RecordError(ctx, *err)
	span.End()
}
