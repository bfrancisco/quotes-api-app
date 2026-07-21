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

func TestInvalidCreateJSONResponse(t *testing.T) {
	response := serve(newRouter(), http.MethodPost, "/v1/quotes", `{"text":`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}

	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if envelope.Error.Code != "INVALID_REQUEST_BODY" {
		t.Fatalf("error code = %q, want %q", envelope.Error.Code, "INVALID_REQUEST_BODY")
	}
}
