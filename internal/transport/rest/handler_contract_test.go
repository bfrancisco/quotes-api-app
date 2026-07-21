package rest_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	testsuite "github.com/bfrancisco/quotes-api-app/internal/transport"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type restHarness struct {
	suite  *suite.Suite
	router *gin.Engine
}

func (h *restHarness) Setup(s *suite.Suite) {
	h.suite = s
	h.router = newRouter()
}

func (h *restHarness) HealthStatus() string {
	response := serve(h.router, http.MethodGet, "/v1/health", "")
	h.suite.Equal(http.StatusOK, response.Code)

	var envelope struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	return envelope.Data.Status
}

func (h *restHarness) CreateQuote(text, author string) string {
	response := serve(h.router, http.MethodPost, "/v1/quotes", `{"text":"`+text+`","author":"`+author+`"}`)
	h.suite.Equal(http.StatusCreated, response.Code)

	var envelope quoteEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	h.suite.NotEmpty(envelope.Data.ID)
	return envelope.Data.ID
}

func (h *restHarness) ListQuotes(author string, limit, offset int) testsuite.QuoteList {
	target := "/v1/quotes?author=" + url.QueryEscape(author) + "&limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	response := serve(h.router, http.MethodGet, target, "")
	h.suite.Equal(http.StatusOK, response.Code)

	var envelope quoteListEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))

	quotes := make([]testsuite.Quote, 0, len(envelope.Data))
	for _, quote := range envelope.Data {
		quotes = append(quotes, testsuite.Quote{ID: quote.ID, Author: quote.Author})
	}

	return testsuite.QuoteList{
		Count:  envelope.Meta.Count,
		Limit:  envelope.Meta.Limit,
		Offset: envelope.Meta.Offset,
		Quotes: quotes,
	}
}

func (h *restHarness) GetQuote(id string) testsuite.Quote {
	response := serve(h.router, http.MethodGet, "/v1/quotes/"+id, "")
	h.suite.Equal(http.StatusOK, response.Code)

	var envelope quoteEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	return testsuite.Quote{ID: envelope.Data.ID, Text: envelope.Data.Text, Author: envelope.Data.Author}
}

func (h *restHarness) RandomQuote() testsuite.Quote {
	response := serve(h.router, http.MethodGet, "/v1/quotes/random", "")
	h.suite.Equal(http.StatusOK, response.Code)

	var envelope quoteEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	return testsuite.Quote{ID: envelope.Data.ID, Text: envelope.Data.Text, Author: envelope.Data.Author}
}

func (h *restHarness) UpdateQuoteText(id, text string) testsuite.Quote {
	response := serve(h.router, http.MethodPatch, "/v1/quotes/"+id, `{"text":"`+text+`"}`)
	h.suite.Equal(http.StatusOK, response.Code)

	var envelope quoteEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	return testsuite.Quote{ID: envelope.Data.ID, Text: envelope.Data.Text, Author: envelope.Data.Author}
}

func (h *restHarness) DeleteQuote(id string) {
	response := serve(h.router, http.MethodDelete, "/v1/quotes/"+id, "")
	h.suite.Equal(http.StatusNoContent, response.Code)
}

func (h *restHarness) ExpectErrorCode(name string) string {
	var method string
	var target string
	var body string
	var wantStatus int

	switch name {
	case "invalid_quote_text":
		method, target, body, wantStatus = http.MethodPost, "/v1/quotes", `{"text":" ","author":"Author"}`, http.StatusBadRequest
	case "invalid_list_options":
		method, target, body, wantStatus = http.MethodGet, "/v1/quotes?limit=101", "", http.StatusBadRequest
	case "invalid_quote_id":
		method, target, body, wantStatus = http.MethodGet, "/v1/quotes/not-a-uuid", "", http.StatusBadRequest
	case "empty_random_store":
		method, target, body, wantStatus = http.MethodGet, "/v1/quotes/random", "", http.StatusNotFound
	case "empty_update":
		method, target, body, wantStatus = http.MethodPatch, "/v1/quotes/550e8400-e29b-41d4-a716-446655440000", `{}`, http.StatusBadRequest
	case "missing_delete":
		method, target, body, wantStatus = http.MethodDelete, "/v1/quotes/550e8400-e29b-41d4-a716-446655440000", "", http.StatusNotFound
	case "missing_get":
		id := h.CreateQuote("temp", "author")
		h.DeleteQuote(id)
		method, target, body, wantStatus = http.MethodGet, "/v1/quotes/"+id, "", http.StatusNotFound
	default:
		h.suite.FailNow("unknown error scenario", name)
	}

	response := serve(h.router, method, target, body)
	h.suite.Equal(wantStatus, response.Code)

	var envelope errorEnvelope
	h.suite.Require().NoError(json.Unmarshal(response.Body.Bytes(), &envelope))
	return envelope.Error.Code
}

func TestRESTTransportContractSuite(t *testing.T) {
	testsuite.RunQuoteTransportContractSuite(t, &restHarness{})
}
