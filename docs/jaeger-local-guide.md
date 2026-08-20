# Local Jaeger tracing guide

This guide runs Jaeger locally through an OpenTelemetry Collector and lets you inspect the traces emitted by the REST and GraphQL APIs.

## Architecture

```text
REST or GraphQL API on the host
  -> OTLP/HTTP http://localhost:4318
  -> OpenTelemetry Collector container
  -> Jaeger container
  -> Jaeger UI http://localhost:16686
```

The applications export **OTLP**, not Jaeger protocol. The Collector owns the backend connection, so replacing Jaeger later does not require application instrumentation changes.

## Prerequisites

- Docker Desktop or Docker Engine is running.
- Go 1.26.4 or later is installed.
- Run commands from the repository root.

> **Proxy note:** If `HTTP_PROXY` or `http_proxy` is set, bypass it for the local API, Collector, and Jaeger endpoints:
>
> ```bash
> export NO_PROXY=localhost,127.0.0.1
> export no_proxy=localhost,127.0.0.1
> ```
>
> Without this, a proxy can intercept local requests and OTLP exports.

## 1. Start Jaeger and the Collector

```bash
docker compose -f docker-compose.observability.yml up -d
```

Confirm both containers are running:

```bash
docker compose -f docker-compose.observability.yml ps
```

Open the Jaeger UI at <http://localhost:16686>.

The host exposes only:

| Port | Service | Purpose |
| --- | --- | --- |
| `16686` | Jaeger | Trace UI |
| `4318` | OpenTelemetry Collector | OTLP/HTTP receiver for locally run APIs |

## 2. Run the REST API with tracing

In a separate terminal:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
export DEPLOYMENT_ENVIRONMENT=local
export OTEL_SERVICE_VERSION=local
SEED_QUOTES=true go run ./cmd/rest-api
```

Create a quote in another terminal:

```bash
curl -i -X POST http://localhost:8080/v1/quotes \
  -H 'Content-Type: application/json' \
  -d '{"text":"Tracing shows the path of a request.","author":"Quotes API"}'
```

The response includes `X-Trace-ID`. Copy that value to locate the trace in Jaeger.

Expected span tree:

```text
POST /v1/quotes
└── quote.create
```

## 3. Run the GraphQL API with tracing

In a separate terminal:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
export DEPLOYMENT_ENVIRONMENT=local
export OTEL_SERVICE_VERSION=local
SEED_QUOTES=true go run ./cmd/graphql-api
```

Run a query:

```bash
curl -i -X POST http://localhost:8081/graphql/query \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ quotes { count } }"}'
```

Expected span tree:

```text
POST /graphql/query
└── graphql.query
    └── quote.list
```

## 4. Find a trace in Jaeger

1. Open <http://localhost:16686>.
2. Select `quotes-rest-api` or `quotes-graphql-api` in **Service**.
3. Choose an operation, then click **Find Traces**.
4. Open a result and inspect the hierarchy and duration.
5. Optionally use the response `X-Trace-ID` in Jaeger's trace-ID lookup when available in the current UI.

Use [telemetry-spans.md](telemetry-spans.md) as the source of truth for expected span names, parents, attributes, error behavior, and privacy constraints.

## 5. Test W3C propagation

Supply a valid W3C trace context header to prove that the HTTP server span joins an upstream trace:

```bash
curl -i http://localhost:8080/v1/quotes \
  -H 'traceparent: 00-11111111111111111111111111111111-2222222222222222-01'
```

The API starts a child span under that trace ID and returns the same trace ID in `X-Trace-ID`.

## 6. Firestore emulator traces

When using the Docker Firestore Emulator, use the same telemetry environment variables for the API. The Firestore client automatically emits client-library spans. The verified create hierarchy is documented in [telemetry-spans.md](telemetry-spans.md).

Do not add quote text, authors, document IDs, query contents, or transaction data as custom attributes.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Jaeger has no service entries | Confirm the API was started with `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318`, make a non-health request, and ensure `NO_PROXY`/`no_proxy` include `localhost,127.0.0.1` when a proxy is configured. |
| API startup cannot create an exporter | Ensure port `4318` is published by the Collector and inspect `docker compose -f docker-compose.observability.yml logs otel-collector`. |
| Collector cannot export | Inspect `docker compose -f docker-compose.observability.yml logs jaeger otel-collector`. |
| Health request does not appear | Expected: `/v1/health` is intentionally untraced. |
| Playground request does not appear | Expected: GraphQL playground traffic is intentionally untraced. |
| Trace is incomplete at shutdown | Stop the API with `Ctrl+C` and allow its bounded telemetry shutdown to flush queued spans. |

## Stop the local stack

```bash
docker compose -f docker-compose.observability.yml down
```

Jaeger trace storage is ephemeral in this local configuration. Stopping the stack removes the displayed traces.
