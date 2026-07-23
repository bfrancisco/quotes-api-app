package repository

import (
	"context"

	"github.com/bfrancisco/quotes-api-app/internal/quotes/model"
)

// QuoteRepository defines the persistence operations required by quote use cases.
// Storage adapters, such as memory and Firestore, implement this interface.
type QuoteRepository interface {
	CreateQuote(ctx context.Context, quote model.QuoteCreateInput) (model.Quote, error)

	ListQuotes(ctx context.Context) ([]model.Quote, error)
	GetQuoteByID(ctx context.Context, id string) (model.Quote, error)
	GetQuotesByAuthor(ctx context.Context, author string) ([]model.Quote, error)
	GetRandomQuote(ctx context.Context) (model.Quote, error)

	UpdateQuote(ctx context.Context, quote model.QuoteUpdateInput) (model.Quote, error)

	DeleteQuote(ctx context.Context, id string) error
}
