package testsuite

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type Quote struct {
	ID        string
	Text      string
	Author    string
	CreatedAt string
}

type QuoteList struct {
	Count  int
	Limit  int
	Offset int
	Quotes []Quote
}

type QuoteTransportHarness interface {
	Setup(s *suite.Suite)
	HealthStatus() string
	CreateQuote(text, author string) string
	ListQuotes(author string, limit, offset int) QuoteList
	GetQuote(id string) Quote
	RandomQuote() Quote
	UpdateQuoteText(id, text string) Quote
	DeleteQuote(id string)
	ExpectErrorCode(name string) string
}

type QuoteTransportContractSuite struct {
	suite.Suite
	harness QuoteTransportHarness
}

func RunQuoteTransportContractSuite(t *testing.T, harness QuoteTransportHarness) {
	suite.Run(t, &QuoteTransportContractSuite{harness: harness})
}

func (s *QuoteTransportContractSuite) SetupTest() {
	s.harness.Setup(&s.Suite)
}

func (s *QuoteTransportContractSuite) TestHealthAndQuoteLifecycle() {
	s.Equal("ok", s.harness.HealthStatus())

	firstID := s.harness.CreateQuote("First quote", "Author A")
	s.harness.CreateQuote("Second quote", "Author A")
	s.harness.CreateQuote("Third quote", "Author B")

	list := s.harness.ListQuotes("Author A", 1, 1)
	s.Len(list.Quotes, 1)
	s.Equal("Author A", list.Quotes[0].Author)
	s.Equal(1, list.Count)
	s.Equal(1, list.Limit)
	s.Equal(1, list.Offset)

	got := s.harness.GetQuote(firstID)
	s.Equal("First quote", got.Text)
	s.Equal("Author A", got.Author)

	random := s.harness.RandomQuote()
	s.NotEmpty(random.ID)

	updated := s.harness.UpdateQuoteText(firstID, "Updated quote")
	s.Equal("Updated quote", updated.Text)
	s.Equal("Author A", updated.Author)

	s.harness.DeleteQuote(firstID)
	s.Equal("QUOTE_NOT_FOUND", s.harness.ExpectErrorCode("missing_get"))
}

func (s *QuoteTransportContractSuite) TestCommonErrorCodes() {
	cases := []struct {
		name string
		want string
	}{
		{name: "invalid_quote_text", want: "INVALID_QUOTE_TEXT"},
		{name: "invalid_list_options", want: "INVALID_QUERY_PARAMS"},
		{name: "invalid_quote_id", want: "INVALID_QUOTE_ID"},
		{name: "empty_random_store", want: "QUOTE_NOT_FOUND"},
		{name: "empty_update", want: "NO_FIELDS_TO_UPDATE"},
		{name: "missing_delete", want: "QUOTE_NOT_FOUND"},
	}

	for _, test := range cases {
		test := test
		s.Run(test.name, func() {
			s.Equal(test.want, s.harness.ExpectErrorCode(test.name))
		})
	}
}
