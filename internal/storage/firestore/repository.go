package firestore

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Repository persists quotes in a Firestore collection.
type Repository struct {
	client     *firestoreclient.Client
	collection *firestoreclient.CollectionRef
	newID      func() string
	now        func() time.Time
}

var _ repository.QuoteRepository = (*Repository)(nil)

// NewRepository creates a repository using collectionName. The caller owns closing client.
func NewRepository(client *firestoreclient.Client, collectionName string) (*Repository, error) {
	if client == nil {
		return nil, fmt.Errorf("Firestore client must not be nil")
	}
	if collectionName == "" {
		return nil, fmt.Errorf("Firestore collection must not be empty")
	}

	return &Repository{
		client:     client,
		collection: client.Collection(collectionName),
		newID:      uuid.NewString,
		now:        time.Now,
	}, nil
}

func (r *Repository) CreateQuote(ctx context.Context, input model.QuoteCreateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	quote := model.Quote{
		ID:        r.newID(),
		Text:      input.Text,
		Author:    input.Author,
		// Firestore stores timestamps with microsecond precision. Normalize before
		// persisting so the quote returned from CreateQuote matches later reads.
		CreatedAt: r.now().UTC().Truncate(time.Microsecond),
	}
	_, err := r.collection.Doc(quote.ID).Create(ctx, quoteDocument{Text: quote.Text, Author: quote.Author, CreatedAt: quote.CreatedAt})
	if err != nil {
		return model.Quote{}, err
	}
	return quote, nil
}

func (r *Repository) ListQuotes(ctx context.Context) ([]model.Quote, error) {
	return r.list(ctx, r.collection.OrderBy("createdAt", firestoreclient.Asc).OrderBy(firestoreclient.DocumentID, firestoreclient.Asc))
}

func (r *Repository) GetQuoteByID(ctx context.Context, id string) (model.Quote, error) {
	snapshot, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return model.Quote{}, mapNotFound(err)
	}
	return quoteFromSnapshot(snapshot)
}

func (r *Repository) GetQuotesByAuthor(ctx context.Context, author string) ([]model.Quote, error) {
	query := r.collection.Where("author", "==", author).OrderBy("createdAt", firestoreclient.Asc).OrderBy(firestoreclient.DocumentID, firestoreclient.Asc)
	return r.list(ctx, query)
}

func (r *Repository) GetRandomQuote(ctx context.Context) (model.Quote, error) {
	// This reads the collection and is intentionally bounded by the small expected quote set.
	// Introduce a random-key field before using this endpoint on a large collection.
	quotes, err := r.ListQuotes(ctx)
	if err != nil {
		return model.Quote{}, err
	}
	if len(quotes) == 0 {
		return model.Quote{}, model.ErrQuoteNotFound
	}
	return quotes[rand.IntN(len(quotes))], nil
}

func (r *Repository) UpdateQuote(ctx context.Context, input model.QuoteUpdateInput) (model.Quote, error) {
	if err := input.Validate(); err != nil {
		return model.Quote{}, err
	}

	var quote model.Quote
	err := r.client.RunTransaction(ctx, func(ctx context.Context, transaction *firestoreclient.Transaction) error {
		snapshot, err := transaction.Get(r.collection.Doc(input.ID))
		if err != nil {
			return mapNotFound(err)
		}
		quote, err = quoteFromSnapshot(snapshot)
		if err != nil {
			return err
		}
		updates := make([]firestoreclient.Update, 0, 2)
		if input.Text != nil {
			quote.Text = *input.Text
			updates = append(updates, firestoreclient.Update{Path: "text", Value: quote.Text})
		}
		if input.Author != nil {
			quote.Author = *input.Author
			updates = append(updates, firestoreclient.Update{Path: "author", Value: quote.Author})
		}
		return transaction.Update(r.collection.Doc(input.ID), updates)
	})
	if err != nil {
		return model.Quote{}, err
	}
	return quote, nil
}

func (r *Repository) DeleteQuote(ctx context.Context, id string) error {
	return r.client.RunTransaction(ctx, func(ctx context.Context, transaction *firestoreclient.Transaction) error {
		if _, err := transaction.Get(r.collection.Doc(id)); err != nil {
			return mapNotFound(err)
		}
		return transaction.Delete(r.collection.Doc(id))
	})
}

type quoteDocument struct {
	Text      string    `firestore:"text"`
	Author    string    `firestore:"author"`
	CreatedAt time.Time `firestore:"createdAt"`
}

func (r *Repository) list(ctx context.Context, query firestoreclient.Query) ([]model.Quote, error) {
	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	quotes := make([]model.Quote, 0, len(snapshots))
	for _, snapshot := range snapshots {
		quote, err := quoteFromSnapshot(snapshot)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func quoteFromSnapshot(snapshot *firestoreclient.DocumentSnapshot) (model.Quote, error) {
	var document quoteDocument
	if err := snapshot.DataTo(&document); err != nil {
		return model.Quote{}, fmt.Errorf("decode quote %q: %w", snapshot.Ref.ID, err)
	}
	return model.Quote{ID: snapshot.Ref.ID, Text: document.Text, Author: document.Author, CreatedAt: document.CreatedAt}, nil
}

func mapNotFound(err error) error {
	if status.Code(err) == codes.NotFound {
		return model.ErrQuoteNotFound
	}
	return err
}
