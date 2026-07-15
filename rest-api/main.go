package main

import (
	"log"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/gin-gonic/gin"
)

func main() {
	// var store Store
	// store = quotes.NewFirestoreStore() // firestore implementation
	// store = quotes.NewMemoryStore() // in-memory implementation. to replace with firestore implementation

	store := quotes.NewMemoryStore() // in-memory implementation. to replace with firestore implementation
	if err := quotes.SeedQuotes(store); err != nil {
		log.Fatalf("failed to seed quotes: %v", err)
	}
	handler := NewHandler(store)

	router := gin.Default()
	v1 := router.Group("/v1")
	handler.RegisterRoutes(v1)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start REST API server: %v", err)
	}
}
