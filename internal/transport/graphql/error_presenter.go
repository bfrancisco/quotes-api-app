package graphql

import (
	"context"
	"errors"

	gqlgraphql "github.com/99designs/gqlgen/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/telemetry"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorPresenter converts domain errors into safe GraphQL errors with stable codes.
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	graphqlError := gqlgraphql.DefaultErrorPresenter(ctx, err)

	details := telemetry.ClassifyError(err)
	if details.Expected {
		graphqlError.Message = details.Message
		if graphqlError.Extensions == nil {
			graphqlError.Extensions = map[string]any{}
		}
		graphqlError.Extensions["code"] = details.Code

		return graphqlError
	}

	var requestError *gqlerror.Error
	if errors.As(err, &requestError) {
		return graphqlError
	}

	graphqlError.Message = details.Message
	if graphqlError.Extensions == nil {
		graphqlError.Extensions = map[string]any{}
	}
	graphqlError.Extensions["code"] = details.Code

	return graphqlError
}
