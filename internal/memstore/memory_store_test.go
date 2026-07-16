package memstore

import (
	"context"
	"errors"
	"testing"

	"github.com/bfrancisco/quotes-api-app/quotes"
)

func memoryStoreStringPtr(value string) *string {
	return &value
}

func validCreateInput(author string) quotes.QuoteCreateInput {
	return quotes.QuoteCreateInput{
		Text:   "Simplicity is prerequisite for reliability.",
		Author: author,
	}
}

func createQuoteAndReturnID(t *testing.T, ctx context.Context, store quotes.Store, input quotes.QuoteCreateInput) string {
	t.Helper()

	created, err := store.CreateQuote(ctx, input)
	if err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	return created.ID
}

func TestMemoryStoreCreateAndGetQuoteByID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	input := validCreateInput("Edsger W. Dijkstra")

	createdID := createQuoteAndReturnID(t, ctx, store, input)

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.ID != createdID {
		t.Fatalf("GetQuoteByID().ID = %q, want %q", got.ID, createdID)
	}
	if got.Text != input.Text || got.Author != input.Author {
		t.Fatalf("GetQuoteByID() = %+v, want text=%q author=%q", got, input.Text, input.Author)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("GetQuoteByID().CreatedAt = zero value, want non-zero time")
	}
}

func TestMemoryStoreCreateQuoteGeneratesUniqueIDs(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	firstInput := validCreateInput("Author A")
	secondInput := validCreateInput("Author B")

	firstID := createQuoteAndReturnID(t, ctx, store, firstInput)
	secondID := createQuoteAndReturnID(t, ctx, store, secondInput)

	first, err := store.GetQuoteByID(ctx, firstID)
	if err != nil {
		t.Fatalf("GetQuoteByID(firstID) error = %v, want nil", err)
	}
	second, err := store.GetQuoteByID(ctx, secondID)
	if err != nil {
		t.Fatalf("GetQuoteByID(secondID) error = %v, want nil", err)
	}

	if first.ID == second.ID {
		t.Fatalf("Generated IDs must be unique, got (%q, %q)", first.ID, second.ID)
	}
}

func TestMemoryStoreGetQuoteByIDNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.GetQuoteByID(ctx, "11111111-1111-1111-1111-111111111111")
	if !errors.Is(err, quotes.ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() error = %v, want %v", err, quotes.ErrQuoteNotFound)
	}
}

func TestMemoryStoreListQuotes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	q1 := validCreateInput("Author A")
	q2 := validCreateInput("Author B")

	firstID := createQuoteAndReturnID(t, ctx, store, q1)
	secondID := createQuoteAndReturnID(t, ctx, store, q2)

	got, err := store.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}

	if len(got) != 2 {
		t.Fatalf("ListQuotes() length = %d, want 2", len(got))
	}

	gotByID := map[string]quotes.Quote{}
	for _, quote := range got {
		gotByID[quote.ID] = quote
	}

	if _, ok := gotByID[firstID]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", firstID)
	}
	if _, ok := gotByID[secondID]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", secondID)
	}
}

func TestMemoryStoreUpdateQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	original := validCreateInput("Author A")

	createdID := createQuoteAndReturnID(t, ctx, store, original)

	updated := quotes.QuoteUpdateInput{
		ID:     createdID,
		Text:   memoryStoreStringPtr("Updated text"),
		Author: memoryStoreStringPtr("Author B"),
	}

	if _, err := store.UpdateQuote(ctx, updated); err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.Text != "Updated text" || got.Author != "Author B" {
		t.Fatalf("Updated quote = %+v, want text=%q author=%q", got, "Updated text", "Author B")
	}
}

func TestMemoryStoreUpdateQuoteNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.UpdateQuote(ctx, quotes.QuoteUpdateInput{
		ID:     "11111111-1111-1111-1111-111111111111",
		Text:   memoryStoreStringPtr("Simplicity is prerequisite for reliability."),
		Author: memoryStoreStringPtr("Missing"),
	})
	if !errors.Is(err, quotes.ErrQuoteNotFound) {
		t.Fatalf("UpdateQuote() error = %v, want %v", err, quotes.ErrQuoteNotFound)
	}
}

func TestMemoryStoreUpdateQuoteTextOnly(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	createdID := createQuoteAndReturnID(t, ctx, store, validCreateInput("Author A"))

	if _, err := store.UpdateQuote(ctx, quotes.QuoteUpdateInput{
		ID:   createdID,
		Text: memoryStoreStringPtr("Updated text only"),
	}); err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.Text != "Updated text only" {
		t.Fatalf("Updated quote text = %q, want %q", got.Text, "Updated text only")
	}
	if got.Author != "Author A" {
		t.Fatalf("Updated quote author = %q, want %q", got.Author, "Author A")
	}
}

func TestMemoryStoreUpdateQuoteAuthorOnly(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	createdID := createQuoteAndReturnID(t, ctx, store, validCreateInput("Author A"))

	_, err := store.UpdateQuote(ctx, quotes.QuoteUpdateInput{
		ID:     createdID,
		Author: memoryStoreStringPtr("Author B"),
	})
	if err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.Author != "Author B" {
		t.Fatalf("Updated quote author = %q, want %q", got.Author, "Author B")
	}
	if got.Text != "Simplicity is prerequisite for reliability." {
		t.Fatalf("Updated quote text = %q, want %q", got.Text, "Simplicity is prerequisite for reliability.")
	}
}

func TestMemoryStoreUpdateQuoteNoFieldsToUpdate(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	createdID := createQuoteAndReturnID(t, ctx, store, validCreateInput("Author A"))

	_, err := store.UpdateQuote(ctx, quotes.QuoteUpdateInput{ID: createdID})
	if !errors.Is(err, quotes.ErrNoFieldsToUpdate) {
		t.Fatalf("UpdateQuote() error = %v, want %v", err, quotes.ErrNoFieldsToUpdate)
	}

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.Author != "Author A" || got.Text != "Simplicity is prerequisite for reliability." {
		t.Fatalf("Quote changed after no-op update = %+v", got)
	}
}

func TestMemoryStoreUpdateQuoteInvalidFieldDoesNotPersist(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	createdID := createQuoteAndReturnID(t, ctx, store, validCreateInput("Author A"))

	_, err := store.UpdateQuote(ctx, quotes.QuoteUpdateInput{
		ID:   createdID,
		Text: memoryStoreStringPtr(" "),
	})
	if !errors.Is(err, quotes.ErrInvalidQuoteText) {
		t.Fatalf("UpdateQuote() error = %v, want %v", err, quotes.ErrInvalidQuoteText)
	}

	got, err := store.GetQuoteByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.Author != "Author A" || got.Text != "Simplicity is prerequisite for reliability." {
		t.Fatalf("Quote changed after invalid update = %+v", got)
	}
}

func TestMemoryStoreDeleteQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	quote := validCreateInput("Author A")

	createdID := createQuoteAndReturnID(t, ctx, store, quote)

	if err := store.DeleteQuote(ctx, createdID); err != nil {
		t.Fatalf("DeleteQuote() error = %v, want nil", err)
	}

	_, err := store.GetQuoteByID(ctx, createdID)
	if !errors.Is(err, quotes.ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() after delete error = %v, want %v", err, quotes.ErrQuoteNotFound)
	}
}

func TestMemoryStoreDeleteQuoteNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	err := store.DeleteQuote(ctx, "11111111-1111-1111-1111-111111111111")
	if !errors.Is(err, quotes.ErrQuoteNotFound) {
		t.Fatalf("DeleteQuote() error = %v, want %v", err, quotes.ErrQuoteNotFound)
	}
}

func TestMemoryStoreGetQuotesByAuthor(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	q1 := validCreateInput("Author A")
	q2 := validCreateInput("Author A")
	q3 := validCreateInput("Author B")

	for _, quote := range []quotes.QuoteCreateInput{q1, q2, q3} {
		if _, err := store.CreateQuote(ctx, quote); err != nil {
			t.Fatalf("CreateQuote() error = %v, want nil", err)
		}
	}

	got, err := store.GetQuotesByAuthor(ctx, "Author A")
	if err != nil {
		t.Fatalf("GetQuotesByAuthor() error = %v, want nil", err)
	}

	if len(got) != 2 {
		t.Fatalf("GetQuotesByAuthor() length = %d, want 2", len(got))
	}

	for _, quote := range got {
		if quote.Author != "Author A" {
			t.Fatalf("GetQuotesByAuthor() returned quote with unexpected author: %+v", quote)
		}
	}
}

func TestMemoryStoreGetRandomQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	q1 := validCreateInput("Author A")
	q2 := validCreateInput("Author B")

	for _, quote := range []quotes.QuoteCreateInput{q1, q2} {
		createQuoteAndReturnID(t, ctx, store, quote)
	}

	quotesInStore, err := store.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}

	validIDs := map[string]struct{}{}
	for _, quote := range quotesInStore {
		validIDs[quote.ID] = struct{}{}
	}

	got, err := store.GetRandomQuote(ctx)
	if err != nil {
		t.Fatalf("GetRandomQuote() error = %v, want nil", err)
	}

	if _, ok := validIDs[got.ID]; !ok {
		t.Fatalf("GetRandomQuote() returned unexpected quote ID %s", got.ID)
	}
}

func TestMemoryStoreGetRandomQuoteEmptyStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.GetRandomQuote(ctx)
	if !errors.Is(err, quotes.ErrQuoteNotFound) {
		t.Fatalf("GetRandomQuote() error = %v, want %v", err, quotes.ErrQuoteNotFound)
	}
}
