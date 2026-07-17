package memory

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/google/uuid"
)

// Repository is an in-memory implementation of repository.QuoteRepository.
type InMemoryRepository struct {
	mu     sync.RWMutex
	quotes map[string]model.Quote
	ids    []string
	newID  func() string
}

var _ repository.QuoteRepository = (*InMemoryRepository)(nil)

// NewRepository creates an empty in-memory quote repository.
func NewInMemoryRepository() repository.QuoteRepository {
	return &InMemoryRepository{
		quotes: make(map[string]model.Quote),
		ids:    make([]string, 0),
		newID:  uuid.NewString,
	}
}

func (r *InMemoryRepository) CreateQuote(_ context.Context, input model.QuoteCreateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	quote := model.Quote{
		ID:        r.newID(),
		Text:      input.Text,
		Author:    input.Author,
		CreatedAt: time.Now(),
	}
	if _, exists := r.quotes[quote.ID]; exists {
		return model.Quote{}, model.ErrQuoteAlreadyExists
	}

	r.quotes[quote.ID] = quote
	r.ids = append(r.ids, quote.ID)
	return quote, nil
}

func (r *InMemoryRepository) ListQuotes(_ context.Context) ([]model.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quotes := make([]model.Quote, 0, len(r.ids))
	for _, id := range r.ids {
		quotes = append(quotes, r.quotes[id])
	}
	return quotes, nil
}

func (r *InMemoryRepository) GetQuoteByID(_ context.Context, id string) (model.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quote, ok := r.quotes[id]
	if !ok {
		return model.Quote{}, model.ErrQuoteNotFound
	}
	return quote, nil
}

func (r *InMemoryRepository) GetQuotesByAuthor(_ context.Context, author string) ([]model.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quotes := make([]model.Quote, 0)
	for _, id := range r.ids {
		quote := r.quotes[id]
		if quote.Author == author {
			quotes = append(quotes, quote)
		}
	}
	return quotes, nil
}

func (r *InMemoryRepository) GetRandomQuote(_ context.Context) (model.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ids) == 0 {
		return model.Quote{}, model.ErrQuoteNotFound
	}
	return r.quotes[r.ids[rand.Intn(len(r.ids))]], nil
}

func (r *InMemoryRepository) UpdateQuote(_ context.Context, input model.QuoteUpdateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	quote, ok := r.quotes[input.ID]
	if !ok {
		return model.Quote{}, model.ErrQuoteNotFound
	}
	if input.Text != nil {
		quote.Text = *input.Text
	}
	if input.Author != nil {
		quote.Author = *input.Author
	}

	r.quotes[input.ID] = quote
	return quote, nil
}

func (r *InMemoryRepository) DeleteQuote(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.quotes[id]; !ok {
		return model.ErrQuoteNotFound
	}

	delete(r.quotes, id)
	for index, existingID := range r.ids {
		if existingID == id {
			r.ids = append(r.ids[:index], r.ids[index+1:]...)
			break
		}
	}
	return nil
}
