package quotes

import (
	"context"
	"errors"
	"testing"
)

func validCreateInput(author string) QuoteCreateInput {
	return QuoteCreateInput{
		Text:   "Simplicity is prerequisite for reliability.",
		Author: author,
	}
}

func TestMemoryStoreCreateAndGetQuoteByID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	input := validCreateInput("Edsger W. Dijkstra")

	if err := store.CreateQuote(ctx, input); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, "1")
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got.ID != "1" {
		t.Fatalf("GetQuoteByID().ID = %q, want %q", got.ID, "1")
	}
	if got.Text != input.Text || got.Author != input.Author {
		t.Fatalf("GetQuoteByID() = %+v, want text=%q author=%q", got, input.Text, input.Author)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("GetQuoteByID().CreatedAt = zero value, want non-zero time")
	}
}

func TestMemoryStoreCreateQuoteGeneratesSequentialIDs(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	firstInput := validCreateInput("Author A")
	secondInput := validCreateInput("Author B")

	if err := store.CreateQuote(ctx, firstInput); err != nil {
		t.Fatalf("CreateQuote() first call error = %v, want nil", err)
	}
	if err := store.CreateQuote(ctx, secondInput); err != nil {
		t.Fatalf("CreateQuote() second call error = %v, want nil", err)
	}

	first, err := store.GetQuoteByID(ctx, "1")
	if err != nil {
		t.Fatalf("GetQuoteByID(1) error = %v, want nil", err)
	}
	second, err := store.GetQuoteByID(ctx, "2")
	if err != nil {
		t.Fatalf("GetQuoteByID(2) error = %v, want nil", err)
	}

	if first.ID != "1" || second.ID != "2" {
		t.Fatalf("Generated IDs = (%q, %q), want (%q, %q)", first.ID, second.ID, "1", "2")
	}
}

func TestMemoryStoreGetQuoteByIDNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.GetQuoteByID(ctx, "404")
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() error = %v, want %v", err, ErrQuoteNotFound)
	}
}

func TestMemoryStoreListQuotes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	q1 := validCreateInput("Author A")
	q2 := validCreateInput("Author B")

	if err := store.CreateQuote(ctx, q1); err != nil {
		t.Fatalf("CreateQuote(q1) error = %v, want nil", err)
	}
	if err := store.CreateQuote(ctx, q2); err != nil {
		t.Fatalf("CreateQuote(q2) error = %v, want nil", err)
	}

	got, err := store.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("ListQuotes() error = %v, want nil", err)
	}

	if len(got) != 2 {
		t.Fatalf("ListQuotes() length = %d, want 2", len(got))
	}

	gotByID := map[string]Quote{}
	for _, quote := range got {
		gotByID[quote.ID] = quote
	}

	if _, ok := gotByID["1"]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", "1")
	}
	if _, ok := gotByID["2"]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", "2")
	}
}

func TestMemoryStoreUpdateQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	original := validCreateInput("Author A")

	if err := store.CreateQuote(ctx, original); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	updated := QuoteUpdateInput{
		ID:     "1",
		Text:   "Updated text",
		Author: "Author B",
	}

	if err := store.UpdateQuote(ctx, updated); err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, "1")
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

	err := store.UpdateQuote(ctx, QuoteUpdateInput{
		ID:     "404",
		Text:   "Simplicity is prerequisite for reliability.",
		Author: "Missing",
	})
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("UpdateQuote() error = %v, want %v", err, ErrQuoteNotFound)
	}
}

func TestMemoryStoreDeleteQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	quote := validCreateInput("Author A")

	if err := store.CreateQuote(ctx, quote); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	if err := store.DeleteQuote(ctx, "1"); err != nil {
		t.Fatalf("DeleteQuote() error = %v, want nil", err)
	}

	_, err := store.GetQuoteByID(ctx, "1")
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("GetQuoteByID() after delete error = %v, want %v", err, ErrQuoteNotFound)
	}
}

func TestMemoryStoreDeleteQuoteNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	err := store.DeleteQuote(ctx, "404")
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("DeleteQuote() error = %v, want %v", err, ErrQuoteNotFound)
	}
}

func TestMemoryStoreGetQuotesByAuthor(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	q1 := validCreateInput("Author A")
	q2 := validCreateInput("Author A")
	q3 := validCreateInput("Author B")

	for _, quote := range []QuoteCreateInput{q1, q2, q3} {
		if err := store.CreateQuote(ctx, quote); err != nil {
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

	for _, quote := range []QuoteCreateInput{q1, q2} {
		if err := store.CreateQuote(ctx, quote); err != nil {
			t.Fatalf("CreateQuote() error = %v, want nil", err)
		}
	}

	got, err := store.GetRandomQuote(ctx)
	if err != nil {
		t.Fatalf("GetRandomQuote() error = %v, want nil", err)
	}

	if got.ID != "1" && got.ID != "2" {
		t.Fatalf("GetRandomQuote() returned unexpected quote ID %s", got.ID)
	}
}

func TestMemoryStoreGetRandomQuoteEmptyStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.GetRandomQuote(ctx)
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("GetRandomQuote() error = %v, want %v", err, ErrQuoteNotFound)
	}
}
