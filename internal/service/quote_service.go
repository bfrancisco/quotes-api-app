package service

import (
	"context"

	"github.com/bfrancisco/quotes-api-app/internal/helpers"
	"github.com/bfrancisco/quotes-api-app/internal/quotes/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
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

func (s *QuoteService) CreateQuote(ctx context.Context, input model.QuoteCreateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	return s.repository.CreateQuote(ctx, input)
}

func (s *QuoteService) ListQuotes(ctx context.Context, input QuoteListInput) (QuoteListResult, error) {
	limit := input.Limit
	if limit == 0 {
		limit = defaultQuoteListLimit
	}
	if input.Offset < 0 || limit < 1 || limit > 100 {
		return QuoteListResult{}, model.ErrInvalidQuoteListOptions
	}

	var (
		quotes []model.Quote
		err    error
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

func (s *QuoteService) GetQuoteByID(ctx context.Context, id string) (model.Quote, error) {
	if !helpers.IsValidUUID(id) {
		return model.Quote{}, model.ErrInvalidQuoteID
	}

	return s.repository.GetQuoteByID(ctx, id)
}

func (s *QuoteService) GetRandomQuote(ctx context.Context) (model.Quote, error) {
	return s.repository.GetRandomQuote(ctx)
}

func (s *QuoteService) UpdateQuote(ctx context.Context, input model.QuoteUpdateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	return s.repository.UpdateQuote(ctx, input)
}

func (s *QuoteService) DeleteQuote(ctx context.Context, id string) error {
	if !helpers.IsValidUUID(id) {
		return model.ErrInvalidQuoteID
	}

	return s.repository.DeleteQuote(ctx, id)
}
