package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/bfrancisco/quotes-api-app/internal/quotes/model"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
)

func stringPtr(value string) *string {
	return &value
}

func createQuote(t *testing.T, repository interface {
	CreateQuote(context.Context, model.QuoteCreateInput) (model.Quote, error)
}, author string) model.Quote {
	t.Helper()

	quote, err := repository.CreateQuote(context.Background(), model.QuoteCreateInput{
		Text:   "Simplicity is prerequisite for reliability.",
		Author: author,
	})
	if err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}
	return quote
}

func TestRepositoryCRUDAndStableListOrder(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewInMemoryRepository()
	first := createQuote(t, repository, "Author A")
	second := createQuote(t, repository, "Author B")

	quotes, err := repository.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}
	if len(quotes) != 2 || quotes[0].ID != first.ID || quotes[1].ID != second.ID {
		t.Fatalf("ListQuotes() = %#v, want insertion order [%q, %q]", quotes, first.ID, second.ID)
	}

	updated, err := repository.UpdateQuote(ctx, model.QuoteUpdateInput{ID: first.ID, Text: stringPtr("Updated text")})
	if err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}
	if updated.Text != "Updated text" || updated.Author != "Author A" {
		t.Fatalf("UpdateQuote() = %#v, want updated text and unchanged author", updated)
	}

	if err := repository.DeleteQuote(ctx, second.ID); err != nil {
		t.Fatalf("DeleteQuote() error = %v, want nil", err)
	}
	_, err = repository.GetQuoteByID(ctx, second.ID)
	if !errors.Is(err, model.ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() error = %v, want %v", err, model.ErrQuoteNotFound)
	}
}

func TestRepositoryValidationDoesNotMutateStorage(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewInMemoryRepository()

	_, err := repository.CreateQuote(ctx, model.QuoteCreateInput{Text: " ", Author: "Author"})
	if !errors.Is(err, model.ErrInvalidQuoteText) {
		t.Fatalf("CreateQuote() error = %v, want %v", err, model.ErrInvalidQuoteText)
	}

	quotes, err := repository.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}
	if len(quotes) != 0 {
		t.Fatalf("ListQuotes() length = %d, want 0 after invalid create", len(quotes))
	}
}

func TestRepositoryFiltersAndSelectsRandomQuote(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewInMemoryRepository()
	first := createQuote(t, repository, "Author A")
	second := createQuote(t, repository, "Author A")
	createQuote(t, repository, "Author B")

	quotes, err := repository.GetQuotesByAuthor(ctx, "Author A")
	if err != nil {
		t.Fatalf("GetQuotesByAuthor() error = %v, want nil", err)
	}
	if len(quotes) != 2 || quotes[0].ID != first.ID || quotes[1].ID != second.ID {
		t.Fatalf("GetQuotesByAuthor() = %#v, want Author A quotes in insertion order", quotes)
	}

	randomQuote, err := repository.GetRandomQuote(ctx)
	if err != nil {
		t.Fatalf("GetRandomQuote() error = %v, want nil", err)
	}
	if randomQuote.ID == "" {
		t.Fatal("GetRandomQuote() returned a quote without an ID")
	}
}

func TestRepositorySupportsConcurrentAccess(t *testing.T) {
	repository := memory.NewInMemoryRepository()
	const quoteCount = 50

	var waitGroup sync.WaitGroup
	errors := make(chan error, quoteCount)
	for index := 0; index < quoteCount; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := repository.CreateQuote(context.Background(), model.QuoteCreateInput{
				Text:   "Simplicity is prerequisite for reliability.",
				Author: "Concurrent author",
			})
			errors <- err
		}()
	}
	waitGroup.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("CreateQuote() error = %v, want nil", err)
		}
	}

	quotes, err := repository.ListQuotes(context.Background())
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}
	if len(quotes) != quoteCount {
		t.Fatalf("ListQuotes() length = %d, want %d", len(quotes), quoteCount)
	}
}
