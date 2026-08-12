# Test Inventory and Roadmap

## Current Tests by Layer

### Model Layer
- File: `internal/model/validation_test.go`
- TestQuoteValidation: validates Quote entity rules (id, text, author, created time).
- TestInputValidation: validates create/update input rules and rejects empty update payload.

### Service Layer
- File: `internal/service/quote_service_test.go`
- TestCreateQuoteValidatesBeforeCallingRepository: invalid create input fails before repository call.
- TestListQuotesFiltersAndPaginates: verifies author filter and limit/offset behavior.
- TestListQuotesUsesDefaultLimitAndRejectsInvalidPaging: verifies default paging and invalid paging rejection.
- TestGetAndDeleteQuoteValidateIDBeforeCallingRepository: invalid UUID is rejected before repository usage.

### Storage Layer
- File: `internal/storage/memory/repository_test.go`
- TestRepositoryCRUDAndStableListOrder: create/list/update/delete behavior and stable insertion order.
- TestRepositoryValidationDoesNotMutateStorage: invalid create does not mutate repository state.
- TestRepositoryFiltersAndSelectsRandomQuote: author filter and random quote retrieval behavior.
- TestRepositorySupportsConcurrentAccess: concurrent create behavior and resulting count consistency.

### Transport Layer (Shared Contract Suite)
- File: `internal/transport/testsuite/quote_transport_suite.go`
- TestHealthAndQuoteLifecycle: shared lifecycle contract for both REST and GraphQL harnesses.
- TestCommonErrorCodes:
  - invalid_quote_text => INVALID_QUOTE_TEXT
  - invalid_list_options => INVALID_QUERY_PARAMS
  - invalid_quote_id => INVALID_QUOTE_ID
  - empty_random_store => QUOTE_NOT_FOUND
  - empty_update => NO_FIELDS_TO_UPDATE
  - missing_delete => QUOTE_NOT_FOUND

### Transport Layer: GraphQL
- File: `internal/transport/graphql/handler_contract_test.go`
- TestGraphQLTransportContractSuite: runs the shared transport suite against GraphQL transport.
- File: internal/transport/graphql/handler_test.go
- TestGraphQLCreatedAtSerialization: verifies createdAt GraphQL serialization is RFC3339Nano.

### Transport Layer: REST
- File: `internal/transport/rest/handler_contract_test.go`
- TestRESTTransportContractSuite: runs the shared transport suite against REST transport.
- File: internal/transport/rest/handler_test.go
- TestInvalidCreateJSONResponse: validates REST-specific malformed JSON handling (INVALID_REQUEST_BODY).

## Transport Parity Matrix

| Behavior | Shared Suite Case | REST | GraphQL |
|---|---|---|---|
| Health status is ok | TestHealthAndQuoteLifecycle | Yes | Yes |
| Create quote success | TestHealthAndQuoteLifecycle | Yes | Yes |
| List filter and paging (author/limit/offset) | TestHealthAndQuoteLifecycle | Yes | Yes |
| Get by id success | TestHealthAndQuoteLifecycle | Yes | Yes |
| Random quote success | TestHealthAndQuoteLifecycle | Yes | Yes |
| Update quote text | TestHealthAndQuoteLifecycle | Yes | Yes |
| Delete quote success | TestHealthAndQuoteLifecycle | Yes | Yes |
| Get deleted quote => QUOTE_NOT_FOUND | TestHealthAndQuoteLifecycle | Yes | Yes |
| Invalid quote text => INVALID_QUOTE_TEXT | TestCommonErrorCodes | Yes | Yes |
| Invalid list options => INVALID_QUERY_PARAMS | TestCommonErrorCodes | Yes | Yes |
| Invalid quote id => INVALID_QUOTE_ID | TestCommonErrorCodes | Yes | Yes |
| Empty random store => QUOTE_NOT_FOUND | TestCommonErrorCodes | Yes | Yes |
| Empty update => NO_FIELDS_TO_UPDATE | TestCommonErrorCodes | Yes | Yes |
| Missing delete => QUOTE_NOT_FOUND | TestCommonErrorCodes | Yes | Yes |
| Malformed request body => INVALID_REQUEST_BODY | Transport-specific | Yes | N/A |
| createdAt serialization (RFC3339Nano) | Transport-specific | N/A | Yes |

## Running Firestore Emulator Tests Locally

The default `go test ./...` command skips Firestore integration tests unless `FIRESTORE_EMULATOR_HOST` is set.

Start the emulator using one of these approaches.

### Option A: Docker

With a running Docker daemon, start the emulator in one terminal and leave it running:

```bash
docker run --rm --name quotes-firestore-emulator \
  -p 127.0.0.1:8085:8085 \
  gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators \
  gcloud beta emulators firestore start --host-port=0.0.0.0:8085
```

### Option B: Standalone Google Cloud CLI

Use an installation created by the [official Google Cloud CLI installer](https://cloud.google.com/sdk/docs/install), then install the emulator component once and start it:

```bash
gcloud components install beta cloud-firestore-emulator
gcloud beta emulators firestore start --host-port=127.0.0.1:8085
```

Package-manager-managed installations of `gcloud` (including Snap installations) cannot install components with `gcloud components install`; use Docker or a standalone CLI installation instead.

After the emulator is running, configure the test process in another terminal and run all tests:

```bash
export FIRESTORE_EMULATOR_HOST=127.0.0.1:8085
export FIRESTORE_PROJECT_ID=quotes-api-emulator
go test ./...
```

Stop the emulator with `Ctrl+C` when testing is complete.

## Planned and Follow-up Roadmap

- Keep contract tests as the default parity gate for transport behavior changes.
- Maintain Firestore repository coverage as an emulator-backed integration test:
  - Keep `go test ./...` usable without local Google Cloud tooling by skipping this test when `FIRESTORE_EMULATOR_HOST` is unset.
  - In CI, start a Firestore emulator and set `FIRESTORE_EMULATOR_HOST=127.0.0.1:8085` so every change runs the Firestore storage contract suite.
  - For local verification of Firestore changes, start the emulator, set `FIRESTORE_EMULATOR_HOST` (and a test `FIRESTORE_PROJECT_ID`), then run `go test ./...`.
  - Keep emulator tests isolated with a unique collection per test and cleanup after each test run.
- Add edge-case parity scenarios to shared suite when requirements expand (for example invalid author or duplicate quote policy), only if behavior should match across transports.
- Keep protocol-specific behavior in local files:
  - REST: HTTP parsing/envelope/status details.
  - GraphQL: response envelope and serialization details.
- Optionally add CI output parsing to assert both contract suites run identical subtest names for stronger parity auditing.
