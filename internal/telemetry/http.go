package telemetry

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader exposes the active trace ID to callers for support and diagnostics.
const TraceIDHeader = "X-Trace-ID"

// TraceIDHandler adds the active trace ID before the wrapped handler can commit its response.
// It expects an upstream OpenTelemetry HTTP middleware to have created the server span.
func TraceIDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		setTraceID(response.Header(), request.Context())
		next.ServeHTTP(response, request)
	})
}

// GinTraceIDMiddleware adds the active trace ID before the route handler writes its response.
// It must be registered after otelgin.Middleware.
func GinTraceIDMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		setTraceID(context.Writer.Header(), context.Request.Context())
		context.Next()
	}
}

func setTraceID(header http.Header, ctx context.Context) {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		header.Set(TraceIDHeader, spanContext.TraceID().String())
	}
}
