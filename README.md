# Quotes API Benchmark

A learning project comparing REST and GraphQL by building the same Quotes API with both approaches, then benchmarking them side by side.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go |
| REST | [Gin](https://github.com/gin-gonic/gin) |
| GraphQL | [gqlgen](https://gqlgen.com/) |
| Database | Cloud Firestore (Firebase Admin SDK for Go) |
| Containers | Docker |
| Frontend | Vanilla HTML/JS |

## Project Structure

```
quotes-benchmark/
├── quotes/              # Shared business logic — single source of truth
│   ├── quote.go         # Quote domain model
│   ├── store.go         # Firestore-backed store (CRUD)
│   └── seed.go          # Shared seed data
├── rest-api/            # Gin REST API (port 8080)
│   ├── main.go
│   └── Dockerfile
├── graphql-api/         # gqlgen GraphQL API (port 8081)
│   ├── server.go
│   ├── graph/
│   │   ├── schema.graphqls
│   │   ├── resolver.go
│   │   └── schema.resolvers.go
│   └── Dockerfile
├── frontend/            # Benchmark comparison UI
│   └── index.html
├── go.mod               # Single module: quotes-benchmark
├── go.sum
└── README.md
```

Both APIs import the shared `quotes` package, so the domain model, storage logic, and seed data live in exactly one place. A single Go module (`quotes-benchmark`) at the repo root makes `quotes-benchmark/quotes` importable from both `rest-api` and `graphql-api` without duplicating code or juggling per-service modules.

## Getting Started

### Prerequisites
- Go 1.21+
- Docker
- curl or Postman (for testing)

### Run the REST API
```bash
go mod tidy
go run ./rest-api
# Listening on http://localhost:8080
```

### Run the GraphQL API
```bash
go mod tidy
go run ./graphql-api
# Listening on http://localhost:8081
# GraphQL Playground at http://localhost:8081/
```

## API Reference

### REST Endpoints (port 8080)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/quotes` | List all quotes |
| GET | `/quotes/:id` | Get a single quote |
| POST | `/quotes` | Create a quote |
| PUT | `/quotes/:id` | Update a quote |
| DELETE | `/quotes/:id` | Delete a quote |

### GraphQL Schema (port 8081)

```graphql
type Quote {
  id: ID!
  text: String!
  author: String!
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

## Database Schema

Data is stored in a single Cloud Firestore collection named `quotes`. Each document uses an auto-generated string ID and has the following shape:

```jsonc
// Collection: quotes
// Document ID: auto-generated string (e.g. "a1B2c3D4e5F6")
{
  "text": "string",        // required
  "author": "string",      // required
  "createdAt": "timestamp" // set on creation, used for ordering
}
```

## Benchmark Comparison

The frontend UI measures and compares:

| Metric | What it shows |
|--------|---------------|
| Response time (ms) | Client-side latency via `performance.now()` |
| Payload size (bytes) | Response body size |
| Over-fetching | Extra data returned that wasn't needed |
| Multi-request cost | REST multiple calls vs GraphQL single query |

### Test Scenarios
1. Fetch all quotes (all fields vs selected fields)
2. Fetch a single quote by ID
3. Create a quote then fetch updated list

## Development Roadmap

- [ ] Phase 1: REST API with in-memory storage
- [ ] Phase 2: GraphQL API with in-memory storage
- [ ] Phase 3: Run both APIs simultaneously with shared seed data
- [ ] Phase 4: Frontend benchmark page
- [ ] Phase 5: Add timing headers and payload comparison
- [ ] Phase 6: Cloud Firestore backend (Firebase Admin SDK)

## Seed Data

Both APIs are initialized with the same quotes:

| Text | Author |
|------|--------|
| Keep it simple, stupid. | Kelly Johnson |
| Talk is cheap. Show me the code. | Linus Torvalds |
| First, solve the problem. Then, write the code. | John Johnson |
| Code is like humor. When you have to explain it, it's bad. | Cory House |
| Make it work, make it right, make it fast. | Kent Beck |
