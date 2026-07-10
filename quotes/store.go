package quotes

import "context"

type Store interface {
	CreateQuote(ctx context.Context, quote QuoteCreateInput) error

	ListQuotes(ctx context.Context) ([]Quote, error)
	GetQuoteByID(ctx context.Context, id string) (Quote, error)
	GetQuotesByAuthor(ctx context.Context, author string) ([]Quote, error)
	GetRandomQuote(ctx context.Context) (Quote, error)

	UpdateQuote(ctx context.Context, quote QuoteUpdateInput) error

	DeleteQuote(ctx context.Context, id string) error
}
