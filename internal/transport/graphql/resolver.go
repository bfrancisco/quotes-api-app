package graphql

import "github.com/bfrancisco/quotes-api-app/internal/service"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	quoteService *service.QuoteService
}

// NewResolver constructs the GraphQL transport with the shared quote service.
func NewResolver(quoteService *service.QuoteService) *Resolver {
	return &Resolver{quoteService: quoteService}
}
