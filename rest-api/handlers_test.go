package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/gin-gonic/gin"
)

type testErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type testQuoteEnvelope struct {
	Data struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		Author string `json:"author"`
	} `json:"data"`
}

type testHealthEnvelope struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

type testQuoteListEnvelope struct {
	Data []struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		Author string `json:"author"`
	} `json:"data"`
	Meta struct {
		Count  int `json:"count"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"meta"`
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := quotes.NewMemoryStore()
	handler := NewHandler(store)

	router := gin.New()
	v1 := router.Group("/v1")
	handler.RegisterRoutes(v1)

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.Code, http.StatusOK)
	}

	var payload testHealthEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Data.Status != "ok" {
		t.Fatalf("health status = %q, want %q", payload.Data.Status, "ok")
	}

}

func TestCreateQuoteEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		contentType    string
		wantStatusCode int
		wantErrorCode  string
	}{
		{
			name:           "returns 201 for valid payload",
			body:           `{"text":"Simplicity is prerequisite for reliability.","author":"Edsger W. Dijkstra"}`,
			contentType:    "application/json",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "returns 400 for invalid json",
			body:           `{"text":`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantErrorCode:  "INVALID_REQUEST_BODY",
		},
		{
			name:           "returns 400 for invalid domain payload",
			body:           `{"text":" ","author":"Edsger W. Dijkstra"}`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantErrorCode:  "INVALID_QUOTE_TEXT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := quotes.NewMemoryStore()
			handler := NewHandler(store)

			router := gin.New()
			v1 := router.Group("/v1")
			handler.RegisterRoutes(v1)

			req := httptest.NewRequest(http.MethodPost, "/v1/quotes", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != tc.wantStatusCode {
				t.Fatalf("status code = %d, want %d", resp.Code, tc.wantStatusCode)
			}

			if tc.wantErrorCode != "" {
				var errPayload testErrorEnvelope
				if err := json.Unmarshal(resp.Body.Bytes(), &errPayload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if errPayload.Error.Code != tc.wantErrorCode {
					t.Fatalf("error code = %q, want %q", errPayload.Error.Code, tc.wantErrorCode)
				}
				return
			}

			var quotePayload testQuoteEnvelope
			if err := json.Unmarshal(resp.Body.Bytes(), &quotePayload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if quotePayload.Data.ID == "" {
				t.Fatalf("created quote id = empty, want non-empty")
			}
			if quotePayload.Data.Text != "Simplicity is prerequisite for reliability." {
				t.Fatalf("created quote text = %q, want %q", quotePayload.Data.Text, "Simplicity is prerequisite for reliability.")
			}
			if quotePayload.Data.Author != "Edsger W. Dijkstra" {
				t.Fatalf("created quote author = %q, want %q", quotePayload.Data.Author, "Edsger W. Dijkstra")
			}

			location := resp.Header().Get("Location")
			if location == "" {
				t.Fatalf("Location header = empty, want non-empty")
			}
		})
	}
}

func TestGetQuotesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 with empty list and default meta", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusOK)
		}

		var payload testQuoteListEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if len(payload.Data) != 0 {
			t.Fatalf("data length = %d, want 0", len(payload.Data))
		}
		if payload.Meta.Count != 0 {
			t.Fatalf("meta.count = %d, want 0", payload.Meta.Count)
		}
		if payload.Meta.Limit != 20 {
			t.Fatalf("meta.limit = %d, want 20", payload.Meta.Limit)
		}
		if payload.Meta.Offset != 0 {
			t.Fatalf("meta.offset = %d, want 0", payload.Meta.Offset)
		}
	})

	t.Run("returns 200 and paginates with limit and offset", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		for _, body := range []string{
			`{"text":"Quote 1","author":"Author A"}`,
			`{"text":"Quote 2","author":"Author B"}`,
			`{"text":"Quote 3","author":"Author C"}`,
		} {
			req := httptest.NewRequest(http.MethodPost, "/v1/quotes", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusCreated {
				t.Fatalf("seed create status = %d, want %d", resp.Code, http.StatusCreated)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes?limit=2&offset=1", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusOK)
		}

		var payload testQuoteListEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if len(payload.Data) != 2 {
			t.Fatalf("data length = %d, want 2", len(payload.Data))
		}
		if payload.Meta.Count != 2 {
			t.Fatalf("meta.count = %d, want 2", payload.Meta.Count)
		}
		if payload.Meta.Limit != 2 {
			t.Fatalf("meta.limit = %d, want 2", payload.Meta.Limit)
		}
		if payload.Meta.Offset != 1 {
			t.Fatalf("meta.offset = %d, want 1", payload.Meta.Offset)
		}
	})

	t.Run("returns 200 and filters by author", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		for _, body := range []string{
			`{"text":"Quote A1","author":"Author A"}`,
			`{"text":"Quote B1","author":"Author B"}`,
			`{"text":"Quote A2","author":"Author A"}`,
		} {
			req := httptest.NewRequest(http.MethodPost, "/v1/quotes", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusCreated {
				t.Fatalf("seed create status = %d, want %d", resp.Code, http.StatusCreated)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes?author=Author%20A", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusOK)
		}

		var payload testQuoteListEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if len(payload.Data) != 2 {
			t.Fatalf("data length = %d, want 2", len(payload.Data))
		}
		for _, quote := range payload.Data {
			if quote.Author != "Author A" {
				t.Fatalf("author = %q, want %q", quote.Author, "Author A")
			}
		}
	})

	t.Run("returns 400 for invalid query params", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes?limit=0&offset=-1", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusBadRequest)
		}

		var errPayload testErrorEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &errPayload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if errPayload.Error.Code != "INVALID_QUERY_PARAMS" {
			t.Fatalf("error code = %q, want %q", errPayload.Error.Code, "INVALID_QUERY_PARAMS")
		}
	})
}

func TestGetQuoteByIDEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 for existing quote", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		createReq := httptest.NewRequest(http.MethodPost, "/v1/quotes", bytes.NewBufferString(`{"text":"One quote","author":"Author One"}`))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		router.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("seed create status = %d, want %d", createResp.Code, http.StatusCreated)
		}

		getReq := httptest.NewRequest(http.MethodGet, "/v1/quotes/1", nil)
		getResp := httptest.NewRecorder()
		router.ServeHTTP(getResp, getReq)

		if getResp.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", getResp.Code, http.StatusOK)
		}

		var payload testQuoteEnvelope
		if err := json.Unmarshal(getResp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if payload.Data.ID != "1" {
			t.Fatalf("quote id = %q, want %q", payload.Data.ID, "1")
		}
		if payload.Data.Text != "One quote" {
			t.Fatalf("quote text = %q, want %q", payload.Data.Text, "One quote")
		}
		if payload.Data.Author != "Author One" {
			t.Fatalf("quote author = %q, want %q", payload.Data.Author, "Author One")
		}
	})

	t.Run("returns 400 for invalid quote id", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusBadRequest)
		}

		var errPayload testErrorEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &errPayload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if errPayload.Error.Code != "INVALID_QUOTE_ID" {
			t.Fatalf("error code = %q, want %q", errPayload.Error.Code, "INVALID_QUOTE_ID")
		}
	})

	t.Run("returns 404 for non-existent quote", func(t *testing.T) {
		store := quotes.NewMemoryStore()
		handler := NewHandler(store)

		router := gin.New()
		v1 := router.Group("/v1")
		handler.RegisterRoutes(v1)

		req := httptest.NewRequest(http.MethodGet, "/v1/quotes/999", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("status code = %d, want %d", resp.Code, http.StatusNotFound)
		}

		var errPayload testErrorEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &errPayload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if errPayload.Error.Code != "QUOTE_NOT_FOUND" {
			t.Fatalf("error code = %q, want %q", errPayload.Error.Code, "QUOTE_NOT_FOUND")
		}
	})
}
