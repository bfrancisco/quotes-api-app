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
