package firestore_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/model"
	storage "github.com/bfrancisco/quotes-api-app/internal/storage/firestore"
	"github.com/google/uuid"
)

func TestRepositoryWithEmulator(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("set FIRESTORE_EMULATOR_HOST to run Firestore emulator integration tests")
	}

	ctx := context.Background()
	client, err := firestoreclient.NewClient(ctx, emulatorProjectID())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	repository, err := storage.NewRepository(client, "quotes_test_"+uuid.NewString())
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}

	if _, err := repository.GetRandomQuote(ctx); !errors.Is(err, model.ErrQuoteNotFound) {
		t.Fatalf("GetRandomQuote() error = %v, want %v", err, model.ErrQuoteNotFound)
	}

	first, err := repository.CreateQuote(ctx, model.QuoteCreateInput{Text: "First quote", Author: "Author A"})
	if err != nil {
		t.Fatalf("CreateQuote(first) error = %v", err)
	}
	time.Sleep(time.Millisecond)
	second, err := repository.CreateQuote(ctx, model.QuoteCreateInput{Text: "Second quote", Author: "Author A"})
	if err != nil {
		t.Fatalf("CreateQuote(second) error = %v", err)
	}
	if _, err := repository.CreateQuote(ctx, model.QuoteCreateInput{Text: "Third quote", Author: "Author B"}); err != nil {
		t.Fatalf("CreateQuote(third) error = %v", err)
	}

	quotes, err := repository.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v", err)
	}
	if len(quotes) != 3 || quotes[0].ID != first.ID || quotes[1].ID != second.ID {
		t.Fatalf("ListQuotes() = %#v, want stable creation order", quotes)
	}

	authorQuotes, err := repository.GetQuotesByAuthor(ctx, "Author A")
	if err != nil {
		t.Fatalf("GetQuotesByAuthor() error = %v", err)
	}
	if len(authorQuotes) != 2 || authorQuotes[0].ID != first.ID || authorQuotes[1].ID != second.ID {
		t.Fatalf("GetQuotesByAuthor() = %#v, want Author A quotes in creation order", authorQuotes)
	}

	updatedText := "Updated quote"
	updated, err := repository.UpdateQuote(ctx, model.QuoteUpdateInput{ID: first.ID, Text: &updatedText})
	if err != nil {
		t.Fatalf("UpdateQuote() error = %v", err)
	}
	if updated.Text != updatedText || updated.Author != "Author A" || !updated.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("UpdateQuote() = %#v, want updated text only", updated)
	}

	if err := repository.DeleteQuote(ctx, second.ID); err != nil {
		t.Fatalf("DeleteQuote() error = %v", err)
	}
	if _, err := repository.GetQuoteByID(ctx, second.ID); !errors.Is(err, model.ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() error = %v, want %v", err, model.ErrQuoteNotFound)
	}
}

func emulatorProjectID() string {
	if projectID := os.Getenv("FIRESTORE_PROJECT_ID"); projectID != "" {
		return projectID
	}
	return "quotes-api-emulator"
}
