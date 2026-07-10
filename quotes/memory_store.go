package quotes

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type MemoryStore struct {
	quotes map[string]Quote
	ids    []string // used to store the order of quote IDs for random selection.
	maxID  int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		quotes: make(map[string]Quote),
		ids:    make([]string, 0),
		maxID:  0, // first quote will have ID "1". Doesn't decrement on delete.
	}
}

func (s *MemoryStore) CreateQuote(ctx context.Context, quote QuoteCreateInput) error {
	s.maxID++
	id := fmt.Sprintf("%d", s.maxID)
	newQuote := Quote{
		ID:        id,
		Text:      quote.Text,
		Author:    quote.Author,
		CreatedAt: time.Now(),
	}
	if _, exists := s.quotes[newQuote.ID]; exists {
		return ErrQuoteAlreadyExists
	}
	s.ids = append(s.ids, newQuote.ID)
	s.quotes[newQuote.ID] = newQuote
	return s.quotes[newQuote.ID].Validate()
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

// O(1)
func (s *MemoryStore) UpdateQuote(ctx context.Context, quote QuoteUpdateInput) error {
	if err := quote.Validate(); err != nil {
		return err
	}

	if _, ok := s.quotes[quote.ID]; !ok {
		return ErrQuoteNotFound
	}

	updatedQuote := s.quotes[quote.ID]
	if quote.Text != nil {
		updatedQuote.Text = *quote.Text
	}
	if quote.Author != nil {
		updatedQuote.Author = *quote.Author
	}

	if err := updatedQuote.Validate(); err != nil {
		return err
	}

	s.quotes[quote.ID] = updatedQuote
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
