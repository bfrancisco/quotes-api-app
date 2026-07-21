package graphql_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	testsuite "github.com/bfrancisco/quotes-api-app/internal/transport"
	"github.com/stretchr/testify/suite"
)

type graphQLHarness struct {
	suite  *suite.Suite
	server http.Handler
}

func (h *graphQLHarness) Setup(s *suite.Suite) {
	h.suite = s
	h.server = newServer()
}

func (h *graphQLHarness) mustResponse(query string) graphQLResponse {
	response := execute(h.suite.T(), h.server, query)
	return response
}

func (h *graphQLHarness) HealthStatus() string {
	response := h.mustResponse(`{ health { status } }`)
	h.suite.Len(response.Errors, 0)

	var data struct {
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	return data.Health.Status
}

func (h *graphQLHarness) CreateQuote(text, author string) string {
	response := h.mustResponse(fmt.Sprintf(`mutation { createQuote(input: { text: %q, author: %q }) { id } }`, text, author))
	h.suite.Len(response.Errors, 0)

	var data struct {
		CreateQuote struct {
			ID string `json:"id"`
		} `json:"createQuote"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	h.suite.NotEmpty(data.CreateQuote.ID)
	return data.CreateQuote.ID
}

func (h *graphQLHarness) ListQuotes(author string, limit, offset int) testsuite.QuoteList {
	response := h.mustResponse(fmt.Sprintf(`{ quotes(author: %q, limit: %d, offset: %d) { count limit offset quotes { id text author createdAt } } }`, author, limit, offset))
	h.suite.Len(response.Errors, 0)

	var data struct {
		Quotes struct {
			Count  int `json:"count"`
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Quotes []struct {
				ID        string `json:"id"`
				Text      string `json:"text"`
				Author    string `json:"author"`
				CreatedAt string `json:"createdAt"`
			} `json:"quotes"`
		} `json:"quotes"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))

	quotes := make([]testsuite.Quote, 0, len(data.Quotes.Quotes))
	for _, quote := range data.Quotes.Quotes {
		quotes = append(quotes, testsuite.Quote{ID: quote.ID, Text: quote.Text, Author: quote.Author, CreatedAt: quote.CreatedAt})
	}

	return testsuite.QuoteList{Count: data.Quotes.Count, Limit: data.Quotes.Limit, Offset: data.Quotes.Offset, Quotes: quotes}
}

func (h *graphQLHarness) GetQuote(id string) testsuite.Quote {
	response := h.mustResponse(fmt.Sprintf(`{ quote(id: %q) { id text author createdAt } }`, id))
	h.suite.Len(response.Errors, 0)

	var data struct {
		Quote struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			Author    string `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"quote"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	return testsuite.Quote{ID: data.Quote.ID, Text: data.Quote.Text, Author: data.Quote.Author, CreatedAt: data.Quote.CreatedAt}
}

func (h *graphQLHarness) RandomQuote() testsuite.Quote {
	response := h.mustResponse(`{ randomQuote { id text author createdAt } }`)
	h.suite.Len(response.Errors, 0)

	var data struct {
		RandomQuote struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			Author    string `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"randomQuote"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	return testsuite.Quote{ID: data.RandomQuote.ID, Text: data.RandomQuote.Text, Author: data.RandomQuote.Author, CreatedAt: data.RandomQuote.CreatedAt}
}

func (h *graphQLHarness) UpdateQuoteText(id, text string) testsuite.Quote {
	response := h.mustResponse(fmt.Sprintf(`mutation { updateQuote(id: %q, input: { text: %q }) { id text author createdAt } }`, id, text))
	h.suite.Len(response.Errors, 0)

	var data struct {
		UpdateQuote struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			Author    string `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"updateQuote"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	return testsuite.Quote{ID: data.UpdateQuote.ID, Text: data.UpdateQuote.Text, Author: data.UpdateQuote.Author, CreatedAt: data.UpdateQuote.CreatedAt}
}

func (h *graphQLHarness) DeleteQuote(id string) {
	response := h.mustResponse(fmt.Sprintf(`mutation { deleteQuote(id: %q) }`, id))
	h.suite.Len(response.Errors, 0)

	var data struct {
		DeleteQuote bool `json:"deleteQuote"`
	}
	h.suite.Require().NoError(json.Unmarshal(response.Data, &data))
	h.suite.True(data.DeleteQuote)
}

func (h *graphQLHarness) ExpectErrorCode(name string) string {
	var query string

	switch name {
	case "invalid_quote_text":
		query = `mutation { createQuote(input: { text: " ", author: "Author" }) { id } }`
	case "invalid_list_options":
		query = `{ quotes(limit: 101) { count } }`
	case "invalid_quote_id":
		query = `{ quote(id: "not-a-uuid") { id } }`
	case "empty_random_store":
		query = `{ randomQuote { id } }`
	case "empty_update":
		query = `mutation { updateQuote(id: "550e8400-e29b-41d4-a716-446655440000", input: {}) { id } }`
	case "missing_delete":
		query = `mutation { deleteQuote(id: "550e8400-e29b-41d4-a716-446655440000") }`
	case "missing_get":
		id := h.CreateQuote("temp", "author")
		h.DeleteQuote(id)
		query = fmt.Sprintf(`{ quote(id: %q) { id } }`, id)
	default:
		h.suite.FailNow("unknown error scenario", name)
	}

	response := h.mustResponse(query)
	h.suite.NotEmpty(response.Errors)
	code, _ := response.Errors[0].Extensions["code"].(string)
	if strings.TrimSpace(code) == "" {
		h.suite.FailNow("missing GraphQL error code", name)
	}
	return code
}

func TestGraphQLTransportContractSuite(t *testing.T) {
	testsuite.RunQuoteTransportContractSuite(t, &graphQLHarness{})
}
