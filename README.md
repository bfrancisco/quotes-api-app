# Quotes API

A Go learning project that:

- Implements **REST** and **GraphQL** services for the same Quotes API.
- Shares a layered application core across both transports, keeping business rules independent of transport and storage technologies.
- Supports **in-memory** and **Firestore** data-storage implementations.
- Uses Docker to package the services and Artifact Registry to store container images.
- Uses Cloud Run to run the services in the cloud.
- Uses a GCP Load Balancer to expose a public endpoint.

## To-do

- [x] Implement a Gin REST API on port `8080`.
- [x] Add the `models → repository → service → storage` application core.
- [x] Add the REST transport adapter and `cmd/rest-api` composition root.
- [x] Add seeded, concurrency-safe in-memory storage.
- [x] Implement quote create, list/filter/paginate, get, random, partial update, and delete use cases.
- [x] Add model, service, storage, and REST transport tests.
- [x] Add a GraphQL API with gqlgen that calls `QuoteService`.
- [x] Implement a Firestore repository adapter.
- [x] Add Dockerfiles.
- [x] Setup GCP products (Artifact Registry, Cloud Run, Firestore, Load Balancer)
- [x] Deploy the APIs to GCP and test E2E flow.
- [ ] Add OpenTelemetry instrumentation and Jaeger tracing for both API services.
- [ ] Connect GraphQL to Apollo GraphOS / Apollo Studio.

## Project Goals

- Build a simple API with Go.
- Learn application layering and request flow.
- Keep business rules independent from REST, GraphQL, and storage technologies.
- Implement equivalent REST and GraphQL capabilities after the shared application core is stable.
- Add persistent storage, containerization, deployment, and GraphQL tooling incrementally.

## Architecture Direction

The current request flow is:

```text
REST handlers / GraphQL resolvers
                ↓
             services
                ↓
     repository interfaces (ports)
                ↓
 storage adapters (memory, later Firestore)
```

Models are shared by the service and repository boundaries. They do not depend on Gin, gqlgen, Firestore, or any other infrastructure package.

### Layer Responsibilities

| Layer | Responsibility | Must not depend on |
|---|---|---|
| Handlers / resolvers | Decode transport requests, call a service, and serialize transport responses | Storage implementations |
| Services | Execute quote use cases, validate business rules, filter and paginate results | Gin, gqlgen, database SDKs |
| Repositories | Define persistence operations required by services | HTTP or GraphQL types |
| Models | Define quote data, input types, and domain errors | Transport and storage technologies |
| Storage | Implement repository interfaces for a concrete data source | Gin and GraphQL packages |

This separation lets REST handlers and GraphQL resolvers use the same `QuoteService`. Adding Firestore should only require a new storage adapter that implements the repository interface.

## Project Structure

```text
quotes-api-app/
├── cmd/
│   ├── graphql-api/                # GraphQL composition root and server startup
│   └── rest-api/                   # REST composition root and server startup
├── internal/
│   ├── model/                      # Quote entity, inputs, validation, domain errors
│   ├── service/                    # Quote use cases and business rules
│   ├── repository/                 # QuoteRepository interface
│   ├── storage/
│   │   └── memory/                 # In-memory repository adapter
│   ├── transport/
│   │   ├── graphql/                # gqlgen schema, resolvers, error mapping, transport tests
│   │   └── rest/                   # Gin router, handlers, REST DTOs, transport tests
│   ├── seeds/                      # Seed data and initialization
│   └── helpers/                    # Shared string and UUID helpers
├── openapi.yaml                    # REST API contract
├── gqlgen.yml                      # gqlgen generation configuration
├── rest-api-curls.md               # REST request examples
├── go.mod
└── README.md
```

## Domain Model

```go
type Quote struct {
    ID        string
    Text      string
    Author    string
    CreatedAt time.Time
}
```

A quote is created with text and author. Updates use partial-update semantics:

- The quote ID is required.
- At least one field must be supplied.
- Supplied text and author values must be non-empty after trimming whitespace.

## Getting Started

### Prerequisites

- Go 1.26.4 or later.
- `curl`, Postman, or another HTTP client.

Docker, Firebase, and Apollo tooling are not required.

### Runtime Configuration

The APIs read configuration from **process environment variables**. Go does not automatically load a `.env` file, so either export variables in the shell or place them before the `go run` command.

For a seeded local memory store:

```bash
SEED_QUOTES=true STORAGE_MODE=memory go run ./cmd/rest-api
```

Alternatively, export values once for the current shell session:

```bash
export SEED_QUOTES=true
export STORAGE_MODE=memory
go run ./cmd/rest-api
```

For a firestore-based store:

```bash
STORAGE_MODE=firestore \ 
SEED_QUOTES=false \
FIRESTORE_PROJECT_ID=<gcp-project-id> \
FIRESTORE_DATABASE_ID=<firestore-db-id> \
FIRESTORE_COLLECTION=<collection-name> \
go run ./cmd/rest-api
```

| Variable | Local default | Purpose |
|---|---|---|
| `PORT` | `8080` REST; `8081` GraphQL | HTTP port. Cloud Run supplies this automatically. |
| `STORAGE_MODE` | `memory` | Selects `memory` for local use or `firestore` for persistent storage. |
| `SEED_QUOTES` | `false` | Set to `true` only to load local memory sample quotes. |
| `FIRESTORE_PROJECT_ID` | — | Required when `STORAGE_MODE=firestore`. |
| `FIRESTORE_DATABASE_ID` | default database | Optional non-default Firestore database ID. |
| `FIRESTORE_COLLECTION` | `quotes` | Firestore collection name. |

For Cloud Run, configure these values on the Cloud Run service rather than using a `.env` file. Do not commit credentials or secrets in `.env`; add that file to `.gitignore` if you use one locally.

### Run the REST API

```bash
go mod tidy
SEED_QUOTES=true go run ./cmd/rest-api
```

The API starts at:

```text
http://localhost:8080
```

The server uses an in-memory store by default. With `SEED_QUOTES=true`, it loads sample data on startup; restarting the server resets all changes.

### Run the GraphQL API

```bash
go mod tidy
SEED_QUOTES=true go run ./cmd/graphql-api
```

The GraphQL server starts on port `8081` with these endpoints:

| URL | Description |
|---|---|
| `http://localhost:8081/` | GraphiQL playground for interactive operations |
| `http://localhost:8081/query` | GraphQL HTTP endpoint |

Like the REST API, the GraphQL server has its own in-memory store. With `SEED_QUOTES=true`, it loads sample data; restarting either process resets only that process's data.

### Regenerate GraphQL Code

After editing [internal/transport/graphql/schema.graphqls](internal/transport/graphql/schema.graphqls), regenerate gqlgen artifacts from the repository root:

```bash
go run github.com/99designs/gqlgen generate
```

### Run the Tests

```bash
go test ./...
```

## REST API Reference

Base URL:

```text
http://localhost:8080/v1
```

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/quotes` | List quotes; supports `author`, `limit`, and `offset` query parameters |
| GET | `/quotes/:id` | Get a quote by ID |
| GET | `/quotes/random` | Get a random quote |
| POST | `/quotes` | Create a quote |
| PATCH | `/quotes/:id` | Partially update a quote |
| DELETE | `/quotes/:id` | Delete a quote |

Single-resource success responses use a `data` envelope. List responses include `data` and `meta` fields. Errors use the following shape:

```json
{
  "error": {
    "code": "INVALID_QUOTE_TEXT",
    "message": "quote text is required"
  }
}
```

### Create a Quote

```http
POST /v1/quotes
Content-Type: application/json

{
  "text": "Simplicity is prerequisite for reliability.",
  "author": "Edsger W. Dijkstra"
}
```

### Filter and Paginate Quotes

```text
GET /v1/quotes?author=Linus%20Torvalds&limit=10&offset=0
```

See [rest-api-curls.md](rest-api-curls.md) for more request examples and [openapi.yaml](openapi.yaml) for the API contract.

## Learning Outcomes

- Writing and testing REST APIs in Go.
- Applying handlers → services → repositories → storage layering.
- Designing application boundaries around interfaces.
- Implementing CRUD use cases with validation and error handling.
- Sharing business logic between REST and GraphQL transports.
- Replacing in-memory storage with a NoSQL adapter.
- Packaging and deploying independently runnable API services.

