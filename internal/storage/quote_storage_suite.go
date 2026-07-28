package testsuite

import (
	"context"
	"testing"
	"time"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/stretchr/testify/suite"
)

// QuoteStorageHarness creates an isolated repository for each contract test.
// Implementations may register cleanup through the supplied suite.
type QuoteStorageHarness interface {
	Setup(s *suite.Suite) repository.QuoteRepository
}

// QuoteStorageContractSuite verifies behavior shared by every QuoteRepository
// implementation.
type QuoteStorageContractSuite struct {
	suite.Suite
	harness    QuoteStorageHarness
	repository repository.QuoteRepository
}

func RunQuoteStorageContractSuite(t *testing.T, harness QuoteStorageHarness) {
	suite.Run(t, &QuoteStorageContractSuite{harness: harness})
}

func (s *QuoteStorageContractSuite) SetupTest() {
	s.repository = s.harness.Setup(&s.Suite)
	s.Require().NotNil(s.repository)
}

func (s *QuoteStorageContractSuite) TestEmptyRepository() {
	ctx := context.Background()

	quotes, err := s.repository.ListQuotes(ctx)
	s.Require().NoError(err)
	s.Empty(quotes)

	_, err = s.repository.GetRandomQuote(ctx)
	s.ErrorIs(err, model.ErrQuoteNotFound)

	_, err = s.repository.GetQuoteByID(ctx, "f0000000-0000-4000-8000-000000000000")
	s.ErrorIs(err, model.ErrQuoteNotFound)

	err = s.repository.DeleteQuote(ctx, "f0000000-0000-4000-8000-000000000000")
	s.ErrorIs(err, model.ErrQuoteNotFound)
}

func (s *QuoteStorageContractSuite) TestQuoteLifecycleAndStableListOrder() {
	ctx := context.Background()
	first := s.createQuote("First quote", "Author A")
	time.Sleep(time.Millisecond)
	second := s.createQuote("Second quote", "Author B")
	time.Sleep(time.Millisecond)
	third := s.createQuote("Third quote", "Author A")

	quotes, err := s.repository.ListQuotes(ctx)
	s.Require().NoError(err)
	s.Equal([]string{first.ID, second.ID, third.ID}, quoteIDs(quotes))

	got, err := s.repository.GetQuoteByID(ctx, first.ID)
	s.Require().NoError(err)
	s.Equal(first, got)

	updatedText := "Updated quote"
	updated, err := s.repository.UpdateQuote(ctx, model.QuoteUpdateInput{ID: first.ID, Text: &updatedText})
	s.Require().NoError(err)
	s.Equal(updatedText, updated.Text)
	s.Equal(first.Author, updated.Author)
	s.Equal(first.CreatedAt, updated.CreatedAt)

	s.Require().NoError(s.repository.DeleteQuote(ctx, second.ID))
	_, err = s.repository.GetQuoteByID(ctx, second.ID)
	s.ErrorIs(err, model.ErrQuoteNotFound)
}

func (s *QuoteStorageContractSuite) TestFiltersAndSelectsRandomQuote() {
	ctx := context.Background()
	first := s.createQuote("First quote", "Author A")
	time.Sleep(time.Millisecond)
	second := s.createQuote("Second quote", "Author A")
	third := s.createQuote("Third quote", "Author B")

	quotes, err := s.repository.GetQuotesByAuthor(ctx, "Author A")
	s.Require().NoError(err)
	s.Equal([]string{first.ID, second.ID}, quoteIDs(quotes))

	randomQuote, err := s.repository.GetRandomQuote(ctx)
	s.Require().NoError(err)
	s.Contains([]string{first.ID, second.ID, third.ID}, randomQuote.ID)
}

func (s *QuoteStorageContractSuite) TestValidationDoesNotMutateStorage() {
	ctx := context.Background()

	_, err := s.repository.CreateQuote(ctx, model.QuoteCreateInput{Text: " ", Author: "Author"})
	s.ErrorIs(err, model.ErrInvalidQuoteText)

	quotes, err := s.repository.ListQuotes(ctx)
	s.Require().NoError(err)
	s.Empty(quotes)

	quote := s.createQuote("A quote", "Author")
	emptyText := " "
	_, err = s.repository.UpdateQuote(ctx, model.QuoteUpdateInput{ID: quote.ID, Text: &emptyText})
	s.ErrorIs(err, model.ErrInvalidQuoteText)

	stored, err := s.repository.GetQuoteByID(ctx, quote.ID)
	s.Require().NoError(err)
	s.Equal(quote, stored)
}

func (s *QuoteStorageContractSuite) createQuote(text, author string) model.Quote {
	s.T().Helper()

	quote, err := s.repository.CreateQuote(context.Background(), model.QuoteCreateInput{
		Text:   text,
		Author: author,
	})
	s.Require().NoError(err)
	return quote
}

func quoteIDs(quotes []model.Quote) []string {
	ids := make([]string, len(quotes))
	for index, quote := range quotes {
		ids[index] = quote.ID
	}
	return ids
}
