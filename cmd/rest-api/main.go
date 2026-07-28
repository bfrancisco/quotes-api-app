package main

import (
	"context"
	"log"
	"net/http"
	"fmt"

	firestoreclient "cloud.google.com/go/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/repository"
	"github.com/bfrancisco/quotes-api-app/internal/runtime"
	"github.com/bfrancisco/quotes-api-app/internal/seeds"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	firestorestorage "github.com/bfrancisco/quotes-api-app/internal/storage/firestore"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	resttransport "github.com/bfrancisco/quotes-api-app/internal/transport/rest"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := runtime.LoadConfig("8080")
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
		fmt.Println("Seeding quotes...")
		if err := seeds.SeedQuotes(quoteService); err != nil {
			log.Fatalf("failed to seed quotes: %v", err)
		}
	}

	router := gin.Default()
	resttransport.NewHandler(quoteService).RegisterRoutes(router.Group("/v1"))
	server := &http.Server{Addr: ":" + config.Port, Handler: router}
	if err := runtime.Serve(server); err != nil {
		log.Fatalf("failed to start REST API server: %v", err)
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
