# Quotes API Benchmark

A Go learning project that compares REST and GraphQL implementations of the same Quotes API. It provides layered REST and gqlgen GraphQL APIs backed by independent in-memory stores. Persistent storage, containers, and benchmarking are future milestones.

## To-do

- [x] Implement a Gin REST API on port `8080`.
- [x] Add the `models → repository → service → storage` application core.
- [x] Add the REST transport adapter and `cmd/rest-api` composition root.
- [x] Add seeded, concurrency-safe in-memory storage.
- [x] Implement quote create, list/filter/paginate, get, random, partial update, and delete use cases.
- [x] Add model, service, storage, and REST transport tests.
- [x] Add a GraphQL API with gqlgen that calls `QuoteService`.
- [ ] Implement a Firestore repository adapter.
- [ ] Build a browser benchmark frontend for REST and GraphQL.
- [ ] Add Dockerfiles and Docker Compose.
- [ ] Deploy the APIs and connect GraphQL to Apollo GraphOS / Apollo Studio.

## Project Goals

- Build a simple API with Go.
- Learn application layering and request flow.
- Keep business rules independent from REST, GraphQL, and storage technologies.
- Implement equivalent REST and GraphQL capabilities after the shared application core is stable.
- Compare both API styles with measurable browser-side benchmarks.
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

### Run the REST API

```bash
go mod tidy
go run ./cmd/rest-api
```

The API starts at:

```text
http://localhost:8080
```

The server initializes the in-memory store with seed data on startup. Restarting the server resets all changes.

### Run the GraphQL API

```bash
go mod tidy
go run ./cmd/graphql-api
```

The GraphQL server starts on port `8081` with these endpoints:

| URL | Description |
|---|---|
| `http://localhost:8081/` | GraphiQL playground for interactive operations |
| `http://localhost:8081/query` | GraphQL HTTP endpoint |

Like the REST API, the GraphQL server has its own seeded in-memory store. Restarting either process resets only that process's data.

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
- Measuring API behavior from a browser client.
- Packaging and deploying independently runnable API services.

