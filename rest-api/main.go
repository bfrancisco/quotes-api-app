package main

import (
	"log"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/gin-gonic/gin"
)

func main() {
	store := quotes.NewMemoryStore()
	handler := NewHandler(store)

	router := gin.Default()
	v1 := router.Group("/v1")
	handler.RegisterRoutes(v1)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start REST API server: %v", err)
	}
}
