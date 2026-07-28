package memory_test

import (
	"context"
	"sync"
	"testing"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	testsuite "github.com/bfrancisco/quotes-api-app/internal/storage"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	"github.com/stretchr/testify/suite"
)

type memoryHarness struct{}

func (memoryHarness) Setup(_ *suite.Suite) repository.QuoteRepository {
	return memory.NewInMemoryRepository()
}

func TestRepositoryContract(t *testing.T) {
	testsuite.RunQuoteStorageContractSuite(t, memoryHarness{})
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
