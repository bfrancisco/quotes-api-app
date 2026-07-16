package memstore

import (
	"context"
	"math/rand"
	"time"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/google/uuid"
)

type memoryStore struct {
	quotes map[string]quotes.Quote
	ids    []string      // used to store the order of quote IDs for random selection.
	newID  func() string // used to generate new IDs for quotes. Defaults to uuid.NewString if not provided.
}

func NewMemoryStore() quotes.Store {
	return &memoryStore{
		quotes: make(map[string]quotes.Quote),
		ids:    make([]string, 0),
		newID:  uuid.NewString,
	}
}

func (s *memoryStore) CreateQuote(ctx context.Context, quote quotes.QuoteCreateInput) (quotes.Quote, error) {
	id := s.newID()
	newQuote := quotes.Quote{
		ID:        id,
		Text:      quote.Text,
		Author:    quote.Author,
		CreatedAt: time.Now(),
	}
	if _, exists := s.quotes[newQuote.ID]; exists {
		return quotes.Quote{}, quotes.ErrQuoteAlreadyExists
	}
	s.ids = append(s.ids, newQuote.ID)
	s.quotes[newQuote.ID] = newQuote
	return newQuote, s.quotes[newQuote.ID].Validate()
}

func (s *memoryStore) ListQuotes(ctx context.Context) ([]quotes.Quote, error) {
	quotes := make([]quotes.Quote, 0, len(s.quotes))

	for _, quote := range s.quotes {
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// O(1)
func (s *memoryStore) GetQuoteByID(ctx context.Context, id string) (quotes.Quote, error) {
	quote, ok := s.quotes[id]
	if !ok {
		return quotes.Quote{}, quotes.ErrQuoteNotFound
	}

	return quote, nil
}

// O(n). Can be optimized, but we don't expect this to be a common operation.
func (s *memoryStore) GetQuotesByAuthor(ctx context.Context, author string) ([]quotes.Quote, error) {
	var results []quotes.Quote

	for _, quote := range s.quotes {
		if quote.Author == author {
			results = append(results, quote)
		}
	}

	return results, nil
}

// O(1)
func (s *memoryStore) GetRandomQuote(ctx context.Context) (quotes.Quote, error) {
	if len(s.ids) == 0 {
		return quotes.Quote{}, quotes.ErrQuoteNotFound
	}

	randomID := s.ids[rand.Intn(len(s.ids))]
	quote, ok := s.quotes[randomID]
	if !ok {
		return quotes.Quote{}, quotes.ErrQuoteNotFound
	}

	return quote, nil
}

// O(1)
func (s *memoryStore) UpdateQuote(ctx context.Context, quote quotes.QuoteUpdateInput) (quotes.Quote, error) {
	if err := quote.Validate(); err != nil {
		return quotes.Quote{}, err
	}

	if _, ok := s.quotes[quote.ID]; !ok {
		return quotes.Quote{}, quotes.ErrQuoteNotFound
	}

	updatedQuote := s.quotes[quote.ID]
	if quote.Text != nil {
		updatedQuote.Text = *quote.Text
	}
	if quote.Author != nil {
		updatedQuote.Author = *quote.Author
	}

	if err := updatedQuote.Validate(); err != nil {
		return quotes.Quote{}, err
	}

	s.quotes[quote.ID] = updatedQuote
	return updatedQuote, nil
}

// O(n). Expensive to delete, but we don't expect this to be a common operation.
func (s *memoryStore) DeleteQuote(ctx context.Context, id string) error {
	if _, ok := s.quotes[id]; !ok {
		return quotes.ErrQuoteNotFound
	}

	delete(s.quotes, id)
	for i, existingID := range s.ids {
		if existingID == id {
			s.ids = append(s.ids[:i], s.ids[i+1:]...)
			break
		}
	}
	return nil
}
