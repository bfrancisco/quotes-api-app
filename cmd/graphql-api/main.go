package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bfrancisco/quotes-api-app/internal/seeds"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	graphqltransport "github.com/bfrancisco/quotes-api-app/internal/transport/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/transport/graphql/generated"
)

func main() {
	repository := memory.NewInMemoryRepository()
	quoteService := service.NewQuoteService(repository)
	if err := seeds.SeedQuotes(quoteService); err != nil {
		log.Fatalf("failed to seed quotes: %v", err)
	}

	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: graphqltransport.NewResolver(quoteService),
	}))
	server.SetErrorPresenter(graphqltransport.ErrorPresenter)

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("Quotes GraphQL Playground", "/query"))
	mux.Handle("/query", server)

	log.Println("GraphQL playground available at http://localhost:8081/")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("failed to start GraphQL API server: %v", err)
	}
}
