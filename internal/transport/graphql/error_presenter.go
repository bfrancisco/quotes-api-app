package graphql

import (
	"context"
	"errors"

	gqlgraphql "github.com/99designs/gqlgen/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorPresenter converts domain errors into safe GraphQL errors with stable codes.
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	graphqlError := gqlgraphql.DefaultErrorPresenter(ctx, err)

	code, message := serviceErrorDetails(err)
	if code != "" {
		graphqlError.Message = message
		if graphqlError.Extensions == nil {
			graphqlError.Extensions = map[string]any{}
		}
		graphqlError.Extensions["code"] = code

		return graphqlError
	}

	var requestError *gqlerror.Error
	if errors.As(err, &requestError) {
		return graphqlError
	}

	graphqlError.Message = "Unexpected error"
	if graphqlError.Extensions == nil {
		graphqlError.Extensions = map[string]any{}
	}
	graphqlError.Extensions["code"] = "INTERNAL_ERROR"

	return graphqlError
}

func serviceErrorDetails(err error) (string, string) {
	switch {
	case errors.Is(err, model.ErrInvalidQuoteListOptions):
		return "INVALID_QUERY_PARAMS", "Invalid query parameters"
	case errors.Is(err, model.ErrInvalidQuoteID):
		return "INVALID_QUOTE_ID", "Invalid quote ID"
	case errors.Is(err, model.ErrInvalidQuoteText):
		return "INVALID_QUOTE_TEXT", "Invalid quote text"
	case errors.Is(err, model.ErrInvalidQuoteAuthor):
		return "INVALID_QUOTE_AUTHOR", "Invalid quote author"
	case errors.Is(err, model.ErrNoFieldsToUpdate):
		return "NO_FIELDS_TO_UPDATE", "No fields to update"
	case errors.Is(err, model.ErrQuoteAlreadyExists):
		return "QUOTE_ALREADY_EXISTS", "Quote already exists"
	case errors.Is(err, model.ErrQuoteNotFound):
		return "QUOTE_NOT_FOUND", "Quote not found"
	default:
		return "", ""
	}
}
