package quotes

import (
	"context"
	"math/rand"
)

type MemoryStore struct {
	quotes map[string]Quote
	ids    []string // used to store the order of quote IDs for random selection.
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		quotes: make(map[string]Quote),
		ids:    make([]string, 0),
	}
}

func (s *MemoryStore) CreateQuote(ctx context.Context, quote Quote) error {
	if _, exists := s.quotes[quote.ID]; exists {
		return ErrQuoteAlreadyExists
	}
	s.ids = append(s.ids, quote.ID)
	s.quotes[quote.ID] = quote
	s.quotes[quote.ID].Validate()
	return nil
}

func (s *MemoryStore) ListQuotes(ctx context.Context) ([]Quote, error) {
	quotes := make([]Quote, 0, len(s.quotes))

	for _, quote := range s.quotes {
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// O(1)
func (s *MemoryStore) GetQuoteByID(ctx context.Context, id string) (Quote, error) {
	quote, ok := s.quotes[id]
	if !ok {
		return Quote{}, ErrQuoteNotFound
	}

	return quote, nil
}

// O(n). Can be optimized, but we don't expect this to be a common operation.
func (s *MemoryStore) GetQuotesByAuthor(ctx context.Context, author string) ([]Quote, error) {
	var results []Quote

	for _, quote := range s.quotes {
		if quote.Author == author {
			results = append(results, quote)
		}
	}

	return results, nil
}

// O(1)
func (s *MemoryStore) GetRandomQuote(ctx context.Context) (Quote, error) {
	if len(s.ids) == 0 {
		return Quote{}, ErrQuoteNotFound
	}

	randomID := s.ids[rand.Intn(len(s.ids))]
	quote, ok := s.quotes[randomID]
	if !ok {
		return Quote{}, ErrQuoteNotFound
	}

	return quote, nil
}

func (s *MemoryStore) UpdateQuote(ctx context.Context, quote Quote) error {
	if _, ok := s.quotes[quote.ID]; !ok {
		return ErrQuoteNotFound
	}

	s.quotes[quote.ID] = quote
	s.quotes[quote.ID].Validate()
	return nil
}

// O(n). Expensive to delete, but we don't expect this to be a common operation.
func (s *MemoryStore) DeleteQuote(ctx context.Context, id string) error {
	if _, ok := s.quotes[id]; !ok {
		return ErrQuoteNotFound
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
