package main

import (
	"log"

	"github.com/bfrancisco/quotes-api-app/internal/quotes/seeds"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	resttransport "github.com/bfrancisco/quotes-api-app/internal/transport/rest"
	"github.com/gin-gonic/gin"
)

func main() {
	repository := memory.NewInMemoryRepository()
	quoteService := service.NewQuoteService(repository)
	if err := seeds.SeedQuotes(quoteService); err != nil {
		log.Fatalf("failed to seed quotes: %v", err)
	}

	router := gin.Default()
	resttransport.NewHandler(quoteService).RegisterRoutes(router.Group("/v1"))
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start REST API server: %v", err)
	}
}
