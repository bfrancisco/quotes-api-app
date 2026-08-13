package main

import (
	"context"
	"log"
	"net/http"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/bfrancisco/quotes-api-app/internal/runtime"
	"github.com/bfrancisco/quotes-api-app/internal/seeds"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	firestorestorage "github.com/bfrancisco/quotes-api-app/internal/storage/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	graphqltransport "github.com/bfrancisco/quotes-api-app/internal/transport/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/transport/graphql/generated"
)

func main() {
	config, err := runtime.LoadConfig("8081")
	if err != nil {
		log.Fatalf("invalid runtime configuration: %v", err)
	}
	repository, closeRepository, err := newRepository(context.Background(), config)
	if err != nil {
		log.Fatalf("create quote repository: %v", err)
	}
	defer closeRepository()

	quoteService := service.NewQuoteService(repository)
	if config.SeedQuotes {
		if err := seeds.SeedQuotes(quoteService); err != nil {
			log.Fatalf("failed to seed quotes: %v", err)
		}
	}

	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: graphqltransport.NewResolver(quoteService),
	}))
	server.SetErrorPresenter(graphqltransport.ErrorPresenter)

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("Quotes GraphQL Playground", "/graphql/query"))
	mux.Handle("/graphql/query", server)

	log.Printf("GraphQL playground available at http://localhost:%s/", config.Port)
	if err := runtime.Serve(&http.Server{Addr: ":" + config.Port, Handler: mux}); err != nil {
		log.Fatalf("failed to start GraphQL API server: %v", err)
	}
}

func newRepository(ctx context.Context, config runtime.Config) (repository.QuoteRepository, func() error, error) {
	if config.StorageMode == runtime.StorageModeMemory {
		return memory.NewInMemoryRepository(), func() error { return nil }, nil
	}

	var (
		client *firestoreclient.Client
		err    error
	)
	if config.FirestoreDatabaseID == "" {
		client, err = firestoreclient.NewClient(ctx, config.FirestoreProjectID)
	} else {
		client, err = firestoreclient.NewClientWithDatabase(ctx, config.FirestoreProjectID, config.FirestoreDatabaseID)
	}
	if err != nil {
		return nil, nil, err
	}
	storage, err := firestorestorage.NewRepository(client, config.FirestoreCollection)
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	return storage, client.Close, nil
}
