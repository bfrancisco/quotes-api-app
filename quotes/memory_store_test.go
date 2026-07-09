package quotes

import (
	"context"
	"errors"
	"testing"
	"time"
)

func validTestQuote(id, author string) Quote {
	return Quote{
		ID:        id,
		Text:      "Simplicity is prerequisite for reliability.",
		Author:    author,
		CreatedAt: time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
	}
}

func TestMemoryStoreCreateAndGetQuoteByID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	quote := validTestQuote("1", "Edsger W. Dijkstra")

	if err := store.CreateQuote(ctx, quote); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, quote.ID)
	if err != nil {
		t.Fatalf("GetQuoteByID() error = %v, want nil", err)
	}

	if got != quote {
		t.Fatalf("GetQuoteByID() = %+v, want %+v", got, quote)
	}
}

func TestMemoryStoreCreateQuoteDuplicateID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	quote := validTestQuote("1", "Edsger W. Dijkstra")

	if err := store.CreateQuote(ctx, quote); err != nil {
		t.Fatalf("CreateQuote() first call error = %v, want nil", err)
	}

	err := store.CreateQuote(ctx, quote)
	if !errors.Is(err, ErrQuoteAlreadyExists) {
		t.Fatalf("CreateQuote() second call error = %v, want %v", err, ErrQuoteAlreadyExists)
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
	q1 := validTestQuote("1", "Author A")
	q2 := validTestQuote("2", "Author B")

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

	if _, ok := gotByID[q1.ID]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", q1.ID)
	}
	if _, ok := gotByID[q2.ID]; !ok {
		t.Fatalf("ListQuotes() missing quote ID %s", q2.ID)
	}
}

func TestMemoryStoreUpdateQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	original := validTestQuote("1", "Author A")

	if err := store.CreateQuote(ctx, original); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	updated := original
	updated.Text = "Updated text"
	updated.Author = "Author B"

	if err := store.UpdateQuote(ctx, updated); err != nil {
		t.Fatalf("UpdateQuote() error = %v, want nil", err)
	}

	got, err := store.GetQuoteByID(ctx, original.ID)
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

	err := store.UpdateQuote(ctx, validTestQuote("404", "Missing"))
	if !errors.Is(err, ErrQuoteNotFound) {
		t.Fatalf("UpdateQuote() error = %v, want %v", err, ErrQuoteNotFound)
	}
}

func TestMemoryStoreDeleteQuote(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	quote := validTestQuote("1", "Author A")

	if err := store.CreateQuote(ctx, quote); err != nil {
		t.Fatalf("CreateQuote() error = %v, want nil", err)
	}

	if err := store.DeleteQuote(ctx, quote.ID); err != nil {
		t.Fatalf("DeleteQuote() error = %v, want nil", err)
	}

	_, err := store.GetQuoteByID(ctx, quote.ID)
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
	q1 := validTestQuote("1", "Author A")
	q2 := validTestQuote("2", "Author A")
	q3 := validTestQuote("3", "Author B")

	for _, quote := range []Quote{q1, q2, q3} {
		if err := store.CreateQuote(ctx, quote); err != nil {
			t.Fatalf("CreateQuote(%s) error = %v, want nil", quote.ID, err)
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
	q1 := validTestQuote("1", "Author A")
	q2 := validTestQuote("2", "Author B")

	for _, quote := range []Quote{q1, q2} {
		if err := store.CreateQuote(ctx, quote); err != nil {
			t.Fatalf("CreateQuote(%s) error = %v, want nil", quote.ID, err)
		}
	}

	got, err := store.GetRandomQuote(ctx)
	if err != nil {
		t.Fatalf("GetRandomQuote() error = %v, want nil", err)
	}

	if got.ID != q1.ID && got.ID != q2.ID {
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
