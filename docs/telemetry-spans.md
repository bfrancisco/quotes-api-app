# Telemetry span catalog

> **Status:** Source of truth for span names, hierarchy, attributes, error semantics, and privacy rules in Quotes API.
>
> Update this document in the same change that introduces, renames, or removes a span. A proposed span must not be implemented until its name, parent, attributes, and error behavior are defined here.

## 1. Scope and conventions

This catalog describes application spans emitted by the REST and GraphQL APIs. The applications export traces through OTLP; a Collector and backend such as Jaeger are consumers of the data, not part of the span contract.

### Naming rules

- Span names are stable, lowercase, and dot-separated where the operation is application-specific.
- HTTP server spans use a stable HTTP method and **route template**, never a concrete path containing a quote UUID.
- GraphQL operation spans include the operation type and, if provided by the client, the declared operation name. They never include the raw GraphQL document.
- Service spans are shared across REST and GraphQL and use `quote.<verb>`.
- Span names must not contain quote text, author names, UUIDs, request bodies, query strings, GraphQL variables, credentials, or error messages.

### Common span kinds

| Boundary | Span kind | Why |
| --- | --- | --- |
| Inbound HTTP | `SERVER` | The API receives and handles an HTTP request. |
| GraphQL operation | `INTERNAL` | gqlgen executes one operation inside the HTTP request. |
| Quote service operation | `INTERNAL` | The application performs one transport-independent use case. |
| Future Firestore operation | `CLIENT` | The application calls an external database service. |

### Resource attributes

Every span carries the process resource configured at startup:

| Attribute | Source | Example |
| --- | --- | --- |
| `service.name` | `OTEL_SERVICE_NAME` or entrypoint default | `quotes-rest-api`, `quotes-graphql-api` |
| `service.version` | `OTEL_SERVICE_VERSION`, if set | release version or Git revision |
| `deployment.environment.name` | `DEPLOYMENT_ENVIRONMENT`, if set | `local`, `staging`, `production` |

## 2. Trace roots and inbound propagation

The W3C `traceparent` header determines whether an incoming request continues an upstream trace or starts a new one.

- A valid inbound `traceparent` becomes the parent of the HTTP server span.
- Without a valid parent, the HTTP server span is a new root span.
- The response includes `X-Trace-ID` when a traced request has a valid active span context.
- The REST health endpoint and GraphQL playground are intentionally excluded from application tracing.

## 3. Current spans

### 3.1 REST HTTP server spans

**Owner:** Gin `otelgin` middleware in `cmd/rest-api`.

| Span name | When emitted | Parent | Kind | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `GET /v1/quotes` | List quotes request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Route template only. |
| `POST /v1/quotes` | Create quote request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Route template only. |
| `GET /v1/quotes/:id` | Get quote request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Never include the actual ID in the name. |
| `GET /v1/quotes/random` | Random quote request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | — |
| `PATCH /v1/quotes/:id` | Update quote request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Never include the actual ID in the name. |
| `DELETE /v1/quotes/:id` | Delete quote request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Never include the actual ID in the name. |

`GET /v1/health` emits **no span** by design.

HTTP instrumentation owns standard HTTP semantic-convention attributes, such as request method, matched route, response status code, server address, protocol, and duration. This project must not manually add duplicate HTTP attributes.

### 3.2 GraphQL HTTP server span

**Owner:** `otelhttp` wrapper in `cmd/graphql-api`.

| Span name | When emitted | Parent | Kind | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `POST /graphql/query` | Every GraphQL operation request | Remote parent or new trace | `SERVER` | Derived by HTTP instrumentation | Stable endpoint name; never use raw query text. |

The GraphQL playground at `/` emits **no application span** by design.

### 3.3 GraphQL operation spans

**Owner:** `OperationTracing` gqlgen extension.

| Span name pattern | When emitted | Parent | Kind | Attributes | Status |
| --- | --- | --- | --- | --- | --- |
| `graphql.query` | Anonymous GraphQL query | `POST /graphql/query` | `INTERNAL` | `graphql.operation.type=query` | Error when gqlgen returns GraphQL errors. |
| `graphql.mutation` | Anonymous GraphQL mutation | `POST /graphql/query` | `INTERNAL` | `graphql.operation.type=mutation` | Error when gqlgen returns GraphQL errors. |
| `graphql.query <name>` | Named GraphQL query | `POST /graphql/query` | `INTERNAL` | `graphql.operation.type=query`, `graphql.operation.name=<name>` | Error when gqlgen returns GraphQL errors. |
| `graphql.mutation <name>` | Named GraphQL mutation | `POST /graphql/query` | `INTERNAL` | `graphql.operation.type=mutation`, `graphql.operation.name=<name>` | Error when gqlgen returns GraphQL errors. |

A GraphQL operation span is one span per operation, **not** one span per field resolver. It is acceptable for `health` to end at this level because it does not call `QuoteService`.

### 3.4 Quote service spans

**Owner:** `QuoteService`.

| Span name | Service method | Parent | Kind | Safe attributes | Expected domain errors |
| --- | --- | --- | --- | --- | --- |
| `quote.create` | `CreateQuote()` | REST server span or GraphQL operation span | `INTERNAL` | None | `INVALID_QUOTE_TEXT`, `INVALID_QUOTE_AUTHOR`, `QUOTE_ALREADY_EXISTS` |
| `quote.list` | `ListQuotes()` | REST server span or GraphQL operation span | `INTERNAL` | `quote.list.limit`, `quote.list.offset`, `quote.list.has_author_filter` | `INVALID_QUERY_PARAMS` |
| `quote.get` | `GetQuoteByID()` | REST server span or GraphQL operation span | `INTERNAL` | None | `INVALID_QUOTE_ID`, `QUOTE_NOT_FOUND` |
| `quote.random` | `GetRandomQuote()` | REST server span or GraphQL operation span | `INTERNAL` | None | `QUOTE_NOT_FOUND` |
| `quote.update` | `UpdateQuote()` | REST server span or GraphQL operation span | `INTERNAL` | None | `INVALID_QUOTE_ID`, `INVALID_QUOTE_TEXT`, `INVALID_QUOTE_AUTHOR`, `NO_FIELDS_TO_UPDATE`, `QUOTE_NOT_FOUND` |
| `quote.delete` | `DeleteQuote()` | REST server span or GraphQL operation span | `INTERNAL` | None | `INVALID_QUOTE_ID`, `QUOTE_NOT_FOUND` |

The service passes the child span context to its repository calls. Any current or future storage span must be a child of the matching `quote.*` span.

## 4. Current hierarchy examples

### REST quote creation

```text
POST /v1/quotes                         [SERVER]
└── quote.create                        [INTERNAL]
```

### GraphQL quote list

```text
POST /graphql/query                     [SERVER]
└── graphql.query                       [INTERNAL]
    └── quote.list                      [INTERNAL]
```

### Named GraphQL mutation

```text
POST /graphql/query                     [SERVER]
└── graphql.mutation UpdateAQuote       [INTERNAL]
    └── quote.update                    [INTERNAL]
```

## 5. Error policy

### Expected domain outcomes

A client-caused, valid business outcome must not look like an infrastructure outage.

For an expected domain error, `telemetry.RecordError()`:

1. sets `error.type` to the stable error code;
2. adds a `domain.error` event with the same `error.type`;
3. leaves the span status unset.

| Error code | Domain condition |
| --- | --- |
| `INVALID_QUERY_PARAMS` | Invalid pagination or list parameters |
| `INVALID_QUOTE_ID` | Invalid UUID |
| `INVALID_QUOTE_TEXT` | Missing or invalid quote text |
| `INVALID_QUOTE_AUTHOR` | Missing or invalid author |
| `NO_FIELDS_TO_UPDATE` | Empty partial update |
| `QUOTE_ALREADY_EXISTS` | Duplicate quote creation |
| `QUOTE_NOT_FOUND` | Requested quote does not exist |

### Unexpected failures

For an unexpected failure, `telemetry.RecordError()`:

1. sets `error.type=INTERNAL_ERROR`;
2. records the Go error on the span;
3. sets span status to `ERROR` with description `INTERNAL_ERROR`.

REST and GraphQL client responses remain sanitized as `INTERNAL_ERROR` / `Unexpected error`.

> **GraphQL note:** the current GraphQL operation interceptor marks an operation span as `ERROR` whenever gqlgen returns GraphQL errors. This includes expected domain errors. A future refinement may align operation-span status with this document's expected-domain-error policy, but must preserve the service-span policy above.

## 6. Attribute allowlist and prohibited data

### Allowed application attributes

| Attribute | Span(s) | Type | Cardinality |
| --- | --- | --- | --- |
| `graphql.operation.type` | `graphql.*` | string | Low: query, mutation, subscription |
| `graphql.operation.name` | Named `graphql.*` | string | Bounded by reviewed API operations; omit anonymous names |
| `quote.list.limit` | `quote.list` | integer | Bounded: 1–100 / default 20 |
| `quote.list.offset` | `quote.list` | integer | Potentially unbounded; retain only while useful and review before high-volume production use |
| `quote.list.has_author_filter` | `quote.list` | boolean | Low |
| `error.type` | spans with errors | string | Bounded to catalog error codes |

### Never record

Do not add any of the following as span names, attributes, event attributes, status descriptions, or metric labels:

- quote IDs or Firestore document IDs;
- quote text;
- author names or filters;
- raw REST bodies, headers, query strings, or authorization data;
- raw GraphQL documents, variables, aliases, or fragments;
- Firestore queries, transaction contents, or document values;
- raw internal error messages or stack traces as manually created attributes;
- collector authentication headers, API keys, or tokens.

## 7. Planned storage spans — not yet implemented

Before explicit Firestore spans are added, validate the Firestore emulator trace output to determine whether existing gRPC instrumentation already produces useful spans. Do not duplicate a useful automatic client span with an application span.

If application-level Firestore spans are required, use only the following catalog entries:

| Planned span name | Repository method | Parent | Kind | Safe attributes | Notes |
| --- | --- | --- | --- | --- | --- |
| `firestore.quote.create` | `CreateQuote()` | `quote.create` | `CLIENT` | `db.system=firestore`, collection name, operation `create` | Do not include document ID or quote values. |
| `firestore.quote.list` | `ListQuotes()` | `quote.list` | `CLIENT` | `db.system=firestore`, collection name, operation `list` | Do not include ordering/query text. |
| `firestore.quote.get` | `GetQuoteByID()` | `quote.get` | `CLIENT` | `db.system=firestore`, collection name, operation `get` | Do not include document ID. |
| `firestore.quote.list_by_author` | `GetQuotesByAuthor()` | `quote.list` | `CLIENT` | `db.system=firestore`, collection name, operation `list_by_author` | Do not include author value. |
| `firestore.quote.random_list` | `GetRandomQuote()` through `ListQuotes()` | `quote.random` | `CLIENT` | `db.system=firestore`, collection name, operation `list` | The random choice itself does not require a database span. |
| `firestore.quote.update` | `UpdateQuote()` transaction | `quote.update` | `CLIENT` | `db.system=firestore`, collection name, operation `update` | One operation span covers retrying transaction work. |
| `firestore.quote.delete` | `DeleteQuote()` transaction | `quote.delete` | `CLIENT` | `db.system=firestore`, collection name, operation `delete` | One operation span covers retrying transaction work. |

When the storage implementation begins, update this document with the verified automatic gRPC span names and attributes. State explicitly whether an application-level Firestore span is emitted, suppressed, or replaced by an automatic span.

## 8. Deliberately absent spans

The following work must not create a span in the current scope:

| Area | Reason |
| --- | --- |
| REST health endpoint | High probe volume and low diagnostic value. |
| GraphQL playground | Static/developer UI traffic, not an API operation. |
| Individual GraphQL field resolvers | Span volume scales with query field count without enough additional diagnostic value. |
| In-memory repository calls | No external I/O; the surrounding `quote.*` span already measures the use case. |
| Quote validation helper calls | Too fine-grained; domain error events on the service span are sufficient. |
| Response serialization | Too fine-grained for current diagnostic needs. |
| Seeding calls | Startup work, not request-path behavior; exclude until startup telemetry has a defined use case. |

## 9. Change checklist

For every new or changed span:

1. Add or update its catalog entry before implementation.
2. Define its name, owner, parent, kind, start/end boundary, safe attributes, error semantics, and privacy constraints.
3. Add automated tests for name, hierarchy, propagation, attributes, and error behavior.
4. Verify that an incoming `traceparent` produces the intended parentage.
5. Inspect local traces to ensure no prohibited data is present.
6. Review new attributes for cardinality and sensitive information.
7. Update the planned/current status in this document.
