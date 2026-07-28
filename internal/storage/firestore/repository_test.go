package firestore_test

import (
	"context"
	"os"
	"testing"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	testsuite "github.com/bfrancisco/quotes-api-app/internal/storage"
	storage "github.com/bfrancisco/quotes-api-app/internal/storage/firestore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type firestoreHarness struct{}

func (firestoreHarness) Setup(s *suite.Suite) repository.QuoteRepository {
	ctx := context.Background()
	client, err := firestoreclient.NewClient(ctx, emulatorProjectID())
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if err := client.Close(); err != nil {
			s.T().Errorf("Close() error = %v", err)
		}
	})

	collectionName := "quotes_test_" + uuid.NewString()
	repository, err := storage.NewRepository(client, collectionName)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if err := deleteCollection(context.Background(), client, collectionName); err != nil {
			s.T().Errorf("deleteCollection() error = %v", err)
		}
	})
	return repository
}

func TestRepositoryContractWithEmulator(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("set FIRESTORE_EMULATOR_HOST to run Firestore emulator integration tests")
	}

	testsuite.RunQuoteStorageContractSuite(t, firestoreHarness{})
}

func emulatorProjectID() string {
	if projectID := os.Getenv("FIRESTORE_PROJECT_ID"); projectID != "" {
		return projectID
	}
	return "quotes-api-emulator"
}

func deleteCollection(ctx context.Context, client *firestoreclient.Client, collectionName string) error {
	snapshots, err := client.Collection(collectionName).Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	for _, snapshot := range snapshots {
		if _, err := snapshot.Ref.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}
