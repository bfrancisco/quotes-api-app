# GraphQL API Usage

The Quotes GraphQL API is implemented with gqlgen and calls the shared `QuoteService`. It is an independently runnable process with its own seeded in-memory store.

## Start the API

From the repository root:

```bash
go run ./cmd/graphql-api
```

| URL | Description |
|---|---|
| `http://localhost:8081/` | GraphiQL playground for interactive operations |
| `http://localhost:8081/query` | GraphQL HTTP endpoint |

The API is seeded when it starts. Restarting the process resets its in-memory data.

## Schema Overview

| Operation | Field | Description |
|---|---|---|
| Query | `health` | Returns API availability status. |
| Query | `quote(id: ID!)` | Gets a quote by UUID. |
| Query | `quotes(author, limit, offset)` | Lists quotes with optional exact-author filtering and offset pagination. |
| Query | `randomQuote` | Gets one random quote. |
| Mutation | `createQuote(input)` | Creates a quote. |
| Mutation | `updateQuote(id, input)` | Partially updates a quote. |
| Mutation | `deleteQuote(id)` | Deletes a quote and returns `true` on success. |

A `Quote` has `id`, `text`, `author`, and `createdAt` fields. `createdAt` uses the `DateTime` scalar and is returned as an RFC3339 timestamp.

`quotes` returns a `QuotePage`:

- `quotes`: the requested page of quotes.
- `count`: number of quotes returned in this page.
- `limit`: effective page size; defaults to 20 and must be from 1 through 100.
- `offset`: effective zero-based offset; must be non-negative.

## Queries

### Health

```graphql
query Health {
  health {
    status
  }
}
```

### List Quotes

```graphql
query ListQuotes {
  quotes(author: "Linus Torvalds", limit: 10, offset: 0) {
    count
    limit
    offset
    quotes {
      id
      text
      author
      createdAt
    }
  }
}
```

Omit `author` to list all quotes. Omit `limit` to use the default of 20, and omit `offset` to start from zero.

### Get a Quote by ID

```graphql
query GetQuote {
  quote(id: "550e8400-e29b-41d4-a716-446655440000") {
    id
    text
    author
    createdAt
  }
}
```

### Get a Random Quote

```graphql
query RandomQuote {
  randomQuote {
    id
    text
    author
    createdAt
  }
}
```

## Mutations

### Create a Quote

```graphql
mutation CreateQuote {
  createQuote(input: {
    text: "Simplicity is prerequisite for reliability."
    author: "Edsger W. Dijkstra"
  }) {
    id
    text
    author
    createdAt
  }
}
```

### Partially Update a Quote

Supply at least one field in `input`. Fields not supplied retain their current values.

```graphql
mutation UpdateQuote {
  updateQuote(
    id: "550e8400-e29b-41d4-a716-446655440000"
    input: { text: "Updated quote text" }
  ) {
    id
    text
    author
    createdAt
  }
}
```

### Delete a Quote

```graphql
mutation DeleteQuote {
  deleteQuote(id: "550e8400-e29b-41d4-a716-446655440000")
}
```

## HTTP Request Example

Send operations to `/query` with a JSON body containing `query`:

```http
POST /query HTTP/1.1
Host: localhost:8081
Content-Type: application/json

{
  "query": "{ health { status } }"
}
```

## Errors

Application failures are returned in GraphQL's `errors` array. Domain failures include a stable `extensions.code` value.

```json
{
  "errors": [
    {
      "message": "Invalid quote ID",
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "path": ["quote"],
      "extensions": {
        "code": "INVALID_QUOTE_ID"
      }
    }
  ],
  "data": null
}
```

| Code | Meaning |
|---|---|
| `INVALID_QUERY_PARAMS` | `limit` or `offset` is outside the permitted range. |
| `INVALID_QUOTE_ID` | The quote ID is not a valid UUID. |
| `INVALID_QUOTE_TEXT` | Quote text is empty or whitespace only. |
| `INVALID_QUOTE_AUTHOR` | Quote author is empty or whitespace only. |
| `NO_FIELDS_TO_UPDATE` | The update input supplies neither `text` nor `author`. |
| `QUOTE_ALREADY_EXISTS` | A conflicting quote already exists. |
| `QUOTE_NOT_FOUND` | The requested quote does not exist, or the store is empty for `randomQuote`. |
| `INTERNAL_ERROR` | An unexpected application failure occurred. |

GraphQL syntax and schema-validation errors retain gqlgen's standard error format and do not use these application codes.

## Regenerate gqlgen Code

After modifying [internal/transport/graphql/schema.graphqls](internal/transport/graphql/schema.graphqls), regenerate gqlgen code from the repository root:

```bash
go run github.com/99designs/gqlgen generate
```

