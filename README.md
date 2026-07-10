# Quotes API Benchmark

A learning project that compares REST and GraphQL by building the same Quotes API with both approaches, then benchmarking them side by side.

The goal is to help a junior software engineer practice backend development with Go while showcasing practical skills in REST, GraphQL, NoSQL persistence, Docker, benchmarking, and GraphQL tooling with Apollo.

## Project Goals

- Build a simple but realistic API using Go.
- Implement the same business features through REST and GraphQL.
- Keep shared business logic in one place to avoid duplication.
- Compare REST and GraphQL using measurable client-side benchmarks.
- Start with in-memory storage, then upgrade to Cloud Firestore.
- Containerize the services for local and deployment-ready workflows.
- Connect the deployed GraphQL API to Apollo GraphOS / Apollo Studio for schema exploration and visibility.

## Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Language | Go | Main backend language |
| REST API | Gin | Lightweight HTTP router for REST endpoints |
| GraphQL API | gqlgen | Type-safe GraphQL server generation for Go |
| Initial Storage | In-memory store | Simple first implementation for learning and testing |
| NoSQL Database | Cloud Firestore | Persistent NoSQL backend after the APIs are working |
| Containers | Docker | Package REST and GraphQL APIs as independent services |
| Frontend | Vanilla HTML/JS | Simple benchmark UI without frontend framework complexity |
| GraphQL Tooling | Apollo GraphOS / Apollo Studio | Schema publishing, Explorer, checks, and observability |

## Why This Stack Fits the Project

Go is a strong choice for learning backend fundamentals because it encourages explicit error handling, clear project structure, and simple concurrency patterns.

Gin keeps the REST API approachable for a first project. It is popular, well-documented, and easy to understand.

gqlgen is more advanced than Gin, but it is a good fit once the REST API is complete. It teaches schema-first GraphQL development, generated types, resolvers, and strongly typed API contracts.

Cloud Firestore is a suitable NoSQL database for this project, but it should not be introduced on day one. Starting with an in-memory store makes the project easier to understand. Firestore can be added later behind the same store interface.

Docker is useful once both APIs are working locally. It should be treated as a deployment and packaging milestone, not an early requirement.

Apollo should be used for GraphQL tooling, not as the primary host for the Go API. The GraphQL API can be deployed to Cloud Run, Render, Fly.io, or another host, then connected to Apollo GraphOS / Apollo Studio.

## Project Structure

```text
quotes-benchmark/
├── quotes/                         # Shared business logic
│   ├── quote.go                    # Quote domain model
│   ├── store.go                    # Store interface
│   ├── memory_store.go             # In-memory implementation
│   ├── firestore_store.go          # Firestore implementation
│   └── seed.go                     # Shared seed data
├── rest-api/                       # Gin REST API, port 8080
│   ├── main.go
│   ├── handlers.go
│   └── Dockerfile
├── graphql-api/                    # gqlgen GraphQL API, port 8081
│   ├── server.go
│   ├── gqlgen.yml
│   ├── graph/
│   │   ├── schema.graphqls
│   │   ├── resolver.go
│   │   └── schema.resolvers.go
│   └── Dockerfile
├── frontend/                       # Benchmark comparison UI
│   └── index.html
├── docker-compose.yml              # Runs both APIs locally
├── go.mod                          # Single module: quotes-benchmark
├── go.sum
└── README.md
```

Both APIs import the shared `quotes` package. The domain model, storage interface, storage implementations, validation, and seed data live in one place.

A single Go module at the repository root makes `quotes-benchmark/quotes` importable from both `rest-api` and `graphql-api` without duplicated code or separate service modules.

## Core Domain Model

```go
type Quote struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}
```

## Store Interface

Both REST and GraphQL should depend on a shared store interface instead of directly depending on Firestore.

```go
type Store interface {
  CreateQuote(ctx context.Context, quote QuoteCreateInput) error

	ListQuotes(ctx context.Context) ([]Quote, error)
  GetQuoteByID(ctx context.Context, id string) (Quote, error)
  GetQuotesByAuthor(ctx context.Context, author string) ([]Quote, error)
  GetRandomQuote(ctx context.Context) (Quote, error)

  UpdateQuote(ctx context.Context, quote QuoteUpdateInput) error

	DeleteQuote(ctx context.Context, id string) error
}
```

Shared input types used by the store:

```go
type QuoteCreateInput struct {
  Text   string
  Author string
}

type QuoteUpdateInput struct {
  ID     string
  Text   *string
  Author *string
}
```

`QuoteUpdateInput` uses partial update semantics:

- `ID` is required.
- At least one of `Text` or `Author` must be provided.
- If `Text` or `Author` is provided, it must be non-empty after trimming.

This allows the project to start with `MemoryStore` and later switch to `FirestoreStore` without rewriting the REST or GraphQL layers.

## Getting Started

### Prerequisites

- Go 1.21+
- Docker
- curl or Postman
- Firebase project, only needed for the Firestore phase

### Run the REST API

```bash
go mod tidy
go run ./rest-api
```

REST API:

```text
http://localhost:8080
```

### Run the GraphQL API

```bash
go mod tidy
go run ./graphql-api
```

GraphQL API:

```text
http://localhost:8081
```

GraphQL Playground or Explorer endpoint:

```text
http://localhost:8081/
```

## REST API Reference

Base URL:

```text
http://localhost:8080
```

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/quotes` | List all quotes |
| GET | `/quotes/:id` | Get one quote |
| POST | `/quotes` | Create a quote |
| PUT | `/quotes/:id` | Update a quote |
| DELETE | `/quotes/:id` | Delete a quote |

### Example REST Create Request

```http
POST /quotes
Content-Type: application/json

{
  "text": "Simplicity is prerequisite for reliability.",
  "author": "Edsger W. Dijkstra"
}
```

## GraphQL API Reference

Base URL:

```text
http://localhost:8081
```

### Schema

```graphql
type Quote {
  id: ID!
  text: String!
  author: String!
  createdAt: String!
}

type Query {
  quotes: [Quote!]!
  quote(id: ID!): Quote
}

type Mutation {
  createQuote(text: String!, author: String!): Quote!
  updateQuote(id: ID!, text: String, author: String): Quote!
  deleteQuote(id: ID!): Boolean!
}
```

### Example GraphQL Query

```graphql
query {
  quotes {
    id
    text
    author
  }
}
```

### Example GraphQL Mutation

```graphql
mutation {
  createQuote(
    text: "Simplicity is prerequisite for reliability."
    author: "Edsger W. Dijkstra"
  ) {
    id
    text
    author
  }
}
```

## Database Schema

Firestore is introduced after the in-memory implementation is complete.

Data is stored in a single Cloud Firestore collection named `quotes`.

```jsonc
// Collection: quotes
// Document ID: auto-generated string
{
  "text": "string",
  "author": "string",
  "createdAt": "timestamp"
}
```

## Benchmark Comparison

The frontend benchmark page compares REST and GraphQL using browser-side measurements.

| Metric | What It Shows |
|--------|---------------|
| Response time | Client-side latency measured with `performance.now()` |
| Payload size | Approximate response body size in bytes |
| Over-fetching | Extra fields returned by REST when fewer fields are needed |
| Multi-request cost | REST multiple calls compared with one GraphQL query |

## Benchmark Scenarios

1. Fetch all quotes.
2. Fetch all quotes with only selected fields.
3. Fetch one quote by ID.
4. Create a quote, then fetch the updated list.
5. Compare REST multiple requests against a single GraphQL query.

## Development Roadmap

### Phase 1: Shared Domain Package

- Define the `Quote` model.
- Define create and update input types.
- Define the shared `Store` interface.
- Add shared seed data.
- Add basic validation for quote text and author.

### Phase 2: REST API with In-Memory Storage

- Build REST endpoints with Gin.
- Use `MemoryStore`.
- Return consistent JSON responses.
- Add basic error handling.
- Test endpoints with curl or Postman.

### Phase 3: REST Tests

- Add handler tests using `httptest`.
- Test list, get, create, update, and delete behavior.
- Test common error cases such as missing quote IDs and invalid input.

### Phase 4: GraphQL API with In-Memory Storage

- Add gqlgen configuration.
- Define the GraphQL schema.
- Generate GraphQL types and resolvers.
- Wire resolvers to the same shared `Store` interface.
- Confirm REST and GraphQL return equivalent quote data.

### Phase 5: Run Both APIs Together

- Run REST on port `8080`.
- Run GraphQL on port `8081`.
- Initialize both with the same seed data.
- Confirm both APIs support the same core operations.

### Phase 6: Benchmark Frontend

- Build a vanilla HTML/JS page.
- Add buttons for each benchmark scenario.
- Display REST and GraphQL results side by side.
- Show response time and payload size.

### Phase 7: Benchmark Enhancements

- Add timing headers from both APIs.
- Add payload size comparison.
- Add over-fetching examples.
- Add multi-request REST comparison against single-query GraphQL.

### Phase 8: Firestore Backend

- Create `FirestoreStore`.
- Store quotes in the `quotes` collection.
- Use Firestore auto-generated document IDs.
- Preserve the same `Store` interface.
- Support local development with the Firestore emulator where possible.

### Phase 9: Docker and Local Orchestration

- Add Dockerfiles for both APIs.
- Add `docker-compose.yml`.
- Run REST, GraphQL, and frontend together locally.
- Document required environment variables.

### Phase 10: Deployment

- Deploy the REST and GraphQL APIs to a cloud host such as Cloud Run, Render, or Fly.io.
- Configure Firestore credentials securely.
- Expose public API URLs.
- Confirm the frontend can call deployed endpoints.

### Phase 11: Apollo GraphOS / Apollo Studio

- Publish the GraphQL schema to Apollo GraphOS.
- Use Apollo Explorer to test GraphQL operations.
- Add schema checks as an optional advanced step.
- Use Apollo Studio for visibility into schema usage and operation behavior.

## Seed Data

| Text | Author |
|------|--------|
| Keep it simple, stupid. | Kelly Johnson |
| Talk is cheap. Show me the code. | Linus Torvalds |
| First, solve the problem. Then, write the code. | John Johnson |
| Code is like humor. When you have to explain it, it's bad. | Cory House |
| Make it work, make it right, make it fast. | Kent Beck |

## Suggested Learning Outcomes

By completing this project, a junior engineer should be able to demonstrate:

- Writing REST APIs in Go.
- Writing GraphQL APIs in Go.
- Sharing business logic across multiple API styles.
- Designing around interfaces.
- Implementing CRUD behavior.
- Using a NoSQL database.
- Measuring basic API performance from the browser.
- Running services with Docker.
- Preparing APIs for deployment.
- Connecting a GraphQL API to Apollo tooling.

## Notes for Beginners

Do not start with Firestore, Docker, or Apollo. Build the in-memory REST API first, then add each layer gradually.

The recommended order is:

1. Make it work in memory.
2. Add REST tests.
3. Add GraphQL.
4. Compare both APIs.
5. Add Firestore.
6. Add Docker.
7. Deploy.
8. Connect GraphQL to Apollo.
