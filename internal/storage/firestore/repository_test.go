package firestore_test

import (
	"context"
	"os"
	"testing"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	testsuite "github.com/bfrancisco/quotes-api-app/internal/storage"
	storage "github.com/bfrancisco/quotes-api-app/internal/storage/firestore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
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

func TestCreateQuoteEmitsAutomaticFirestoreSpansWithEmulator(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("set FIRESTORE_EMULATOR_HOST to run Firestore emulator trace validation")
	}

	ctx := context.Background()
	client, err := firestoreclient.NewClient(ctx, emulatorProjectID())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	collectionName := "quotes_trace_test_" + uuid.NewString()
	repository, err := storage.NewRepository(client, collectionName)
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = deleteCollection(context.Background(), client, collectionName) })

	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	parentContext, parent := provider.Tracer("firestore-trace-validation").Start(ctx, "validation.parent")
	_, err = repository.CreateQuote(parentContext, model.QuoteCreateInput{Text: "private trace test quote", Author: "private trace test author"})
	parent.End()
	if err != nil {
		t.Fatalf("CreateQuote() error = %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 3 {
		t.Fatalf("exported span count = %d, want parent and two automatic Firestore spans: %v", len(spans), spans)
	}
	documentCreateSpan := findSpan(t, spans, "cloud.google.com/go/firestore.DocumentRef.Create")
	commitSpan := findSpan(t, spans, "cloud.google.com/go/firestore.Client.commit")
	if documentCreateSpan.SpanKind != trace.SpanKindInternal || commitSpan.SpanKind != trace.SpanKindInternal {
		t.Fatalf("automatic span kinds = (%s, %s), want INTERNAL", documentCreateSpan.SpanKind, commitSpan.SpanKind)
	}
	if documentCreateSpan.Parent.SpanID() != parent.SpanContext().SpanID() {
		t.Fatalf("document-create span parent = %v, want %v", documentCreateSpan.Parent, parent.SpanContext())
	}
	if commitSpan.Parent.SpanID() != documentCreateSpan.SpanContext.SpanID() {
		t.Fatalf("commit span parent = %v, want document-create span %v", commitSpan.Parent, documentCreateSpan.SpanContext)
	}
	if hasAttributeValue(documentCreateSpan.Attributes, "private trace test quote") ||
		hasAttributeValue(documentCreateSpan.Attributes, "private trace test author") ||
		hasAttributeValue(commitSpan.Attributes, "private trace test quote") ||
		hasAttributeValue(commitSpan.Attributes, "private trace test author") {
		t.Fatalf("automatic Firestore attributes = (%v, %v), must not include quote data", documentCreateSpan.Attributes, commitSpan.Attributes)
	}
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

func findSpan(t *testing.T, spans tracetest.SpanStubs, name string) tracetest.SpanStub {
	t.Helper()
	for _, span := range spans {
		if span.Name == name {
			return span
		}
	}
	t.Fatalf("span %q not found in %v", name, spans)
	return tracetest.SpanStub{}
}

func hasAttributeValue(attributes []attribute.KeyValue, value string) bool {
	for _, attribute := range attributes {
		if attribute.Value.AsString() == value {
			return true
		}
	}
	return false
}
