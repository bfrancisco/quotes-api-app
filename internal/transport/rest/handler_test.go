package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/bfrancisco/quotes-api-app/internal/storage/memory"
	resttransport "github.com/bfrancisco/quotes-api-app/internal/transport/rest"
	"github.com/gin-gonic/gin"
)

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type quoteEnvelope struct {
	Data struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		Author string `json:"author"`
	} `json:"data"`
}

type quoteListEnvelope struct {
	Data []struct {
		ID     string `json:"id"`
		Author string `json:"author"`
	} `json:"data"`
	Meta struct {
		Count  int `json:"count"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"meta"`
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	quoteService := service.NewQuoteService(memory.NewInMemoryRepository())
	resttransport.NewHandler(quoteService).RegisterRoutes(router.Group("/v1"))
	return router
}

func serve(router *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func createQuote(t *testing.T, router *gin.Engine, text, author string) string {
	t.Helper()
	response := serve(router, http.MethodPost, "/v1/quotes", `{"text":"`+text+`","author":"`+author+`"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var envelope quoteEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if envelope.Data.ID == "" {
		t.Fatal("created quote ID is empty")
	}
	return envelope.Data.ID
}

func TestHealthAndQuoteLifecycle(t *testing.T) {
	router := newRouter()

	health := serve(router, http.MethodGet, "/v1/health", "")
	if health.Code != http.StatusOK || health.Body.String() != `{"data":{"status":"ok"}}` {
		t.Fatalf("health response = (%d, %s), want 200 health envelope", health.Code, health.Body.String())
	}

	firstID := createQuote(t, router, "First quote", "Author A")
	createQuote(t, router, "Second quote", "Author A")
	createQuote(t, router, "Third quote", "Author B")

	list := serve(router, http.MethodGet, "/v1/quotes?author=Author%20A&limit=1&offset=1", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.Code, http.StatusOK)
	}
	var listEnvelope quoteListEnvelope
	if err := json.Unmarshal(list.Body.Bytes(), &listEnvelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(listEnvelope.Data) != 1 || listEnvelope.Data[0].Author != "Author A" || listEnvelope.Meta.Count != 1 || listEnvelope.Meta.Limit != 1 || listEnvelope.Meta.Offset != 1 {
		t.Fatalf("list response = %+v, want one paginated quote by Author A", listEnvelope)
	}

	get := serve(router, http.MethodGet, "/v1/quotes/"+firstID, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", get.Code, http.StatusOK)
	}

	random := serve(router, http.MethodGet, "/v1/quotes/random", "")
	if random.Code != http.StatusOK {
		t.Fatalf("random status = %d, want %d", random.Code, http.StatusOK)
	}

	update := serve(router, http.MethodPatch, "/v1/quotes/"+firstID, `{"text":"Updated quote"}`)
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", update.Code, http.StatusOK, update.Body.String())
	}
	var updated quoteEnvelope
	if err := json.Unmarshal(update.Body.Bytes(), &updated); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if updated.Data.Text != "Updated quote" || updated.Data.Author != "Author A" {
		t.Fatalf("updated quote = %+v, want updated text and original author", updated.Data)
	}

	deleteResponse := serve(router, http.MethodDelete, "/v1/quotes/"+firstID, "")
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResponse.Code, http.StatusNoContent)
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		status    int
		errorCode string
	}{
		{"invalid create JSON", http.MethodPost, "/v1/quotes", `{"text":`, http.StatusBadRequest, "INVALID_REQUEST_BODY"},
		{"invalid quote text", http.MethodPost, "/v1/quotes", `{"text":" ","author":"Author"}`, http.StatusBadRequest, "INVALID_QUOTE_TEXT"},
		{"invalid list options", http.MethodGet, "/v1/quotes?limit=101", "", http.StatusBadRequest, "INVALID_QUERY_PARAMS"},
		{"invalid quote ID", http.MethodGet, "/v1/quotes/not-a-uuid", "", http.StatusBadRequest, "INVALID_QUOTE_ID"},
		{"empty random store", http.MethodGet, "/v1/quotes/random", "", http.StatusNotFound, "QUOTE_NOT_FOUND"},
		{"empty update", http.MethodPatch, "/v1/quotes/550e8400-e29b-41d4-a716-446655440000", `{}`, http.StatusBadRequest, "NO_FIELDS_TO_UPDATE"},
		{"missing delete", http.MethodDelete, "/v1/quotes/550e8400-e29b-41d4-a716-446655440000", "", http.StatusNotFound, "QUOTE_NOT_FOUND"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serve(newRouter(), test.method, test.target, test.body)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, test.status, response.Body.String())
			}
			var envelope errorEnvelope
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if envelope.Error.Code != test.errorCode {
				t.Fatalf("error code = %q, want %q", envelope.Error.Code, test.errorCode)
			}
		})
	}
}
