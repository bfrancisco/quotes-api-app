package graphql_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	graphqltransport "github.com/bfrancisco/quotes-api-app/internal/transport/graphql"
	"github.com/bfrancisco/quotes-api-app/internal/transport/graphql/generated"
)

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func newServer() http.Handler {
	quoteService := service.NewQuoteService(memory.NewInMemoryRepository())
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: graphqltransport.NewResolver(quoteService),
	}))
	server.SetErrorPresenter(graphqltransport.ErrorPresenter)
	return server
}

func execute(t *testing.T, server http.Handler, query string) graphQLResponse {
	t.Helper()

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var result graphQLResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return result
}

func assertNoErrors(t *testing.T, response graphQLResponse) {
	t.Helper()
	if len(response.Errors) != 0 {
		t.Fatalf("errors = %+v, want none", response.Errors)
	}
}

func assertErrorCode(t *testing.T, response graphQLResponse, want string) {
	t.Helper()
	if len(response.Errors) != 1 {
		t.Fatalf("error count = %d, want 1", len(response.Errors))
	}
	if code, _ := response.Errors[0].Extensions["code"].(string); code != want {
		t.Fatalf("error code = %q, want %q", code, want)
	}
}

func createQuote(t *testing.T, server http.Handler, text, author string) string {
	t.Helper()

	response := execute(t, server, `mutation { createQuote(input: { text: "`+text+`", author: "`+author+`" }) { id } }`)
	assertNoErrors(t, response)

	var data struct {
		CreateQuote struct {
			ID string `json:"id"`
		} `json:"createQuote"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if data.CreateQuote.ID == "" {
		t.Fatal("created quote ID is empty")
	}
	return data.CreateQuote.ID
}

func TestGraphQLCreatedAtSerialization(t *testing.T) {
	server := newServer()
	id := createQuote(t, server, "First quote", "Author A")

	get := execute(t, server, `{ quote(id: "`+id+`") { createdAt } }`)
	assertNoErrors(t, get)

	var data struct {
		Quote struct {
			CreatedAt string `json:"createdAt"`
		} `json:"quote"`
	}
	if err := json.Unmarshal(get.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if _, err := time.Parse(time.RFC3339Nano, data.Quote.CreatedAt); err != nil {
		t.Fatalf("createdAt = %q, want RFC3339 timestamp: %v", data.Quote.CreatedAt, err)
	}
}
