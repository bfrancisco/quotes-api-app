package graphql

import (
	"context"

	gqlgraphql "github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OperationTracing creates one span for each GraphQL operation. It intentionally records only
// operation type and name, never raw operation text or variables.
type OperationTracing struct{}

var _ gqlgraphql.HandlerExtension = OperationTracing{}
var _ gqlgraphql.OperationInterceptor = OperationTracing{}

func (OperationTracing) ExtensionName() string {
	return "OperationTracing"
}

func (OperationTracing) Validate(gqlgraphql.ExecutableSchema) error {
	return nil
}

func (OperationTracing) InterceptOperation(ctx context.Context, next gqlgraphql.OperationHandler) gqlgraphql.ResponseHandler {
	operationContext := gqlgraphql.GetOperationContext(ctx)
	operationType := string(operationContext.Operation.Operation)
	operationName := operationContext.OperationName
	spanName := "graphql." + operationType
	if operationName != "" {
		spanName += " " + operationName
	}

	ctx, span := otel.Tracer("quotes-api/graphql").Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String("graphql.operation.type", operationType))
	if operationName != "" {
		span.SetAttributes(attribute.String("graphql.operation.name", operationName))
	}

	nextHandler := next(ctx)
	return func(responseContext context.Context) *gqlgraphql.Response {
		response := nextHandler(responseContext)
		if response != nil && response.Errors != nil {
			span.SetStatus(codes.Error, "GraphQL operation returned errors")
		}
		span.End()
		return response
	}
}
