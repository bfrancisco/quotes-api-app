package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/bfrancisco/quotes-api-app/internal/model"
	"github.com/bfrancisco/quotes-api-app/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler adapts HTTP requests and responses to quote application services.
type Handler struct {
	service *service.QuoteService
}

type createQuoteRequest struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

type getQuotesRequest struct {
	Author string `form:"author"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type updateQuoteRequest struct {
	Text   *string `json:"text"`
	Author *string `json:"author"`
}

type quoteListMeta struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type quoteListResponse struct {
	Data []quotePayload `json:"data"`
	Meta quoteListMeta  `json:"meta"`
}

type quotePayload struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

type quoteResponse struct {
	Data quotePayload `json:"data"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type healthResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

// NewHandler constructs the REST transport with the shared quote service.
func NewHandler(service *service.QuoteService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches all version-one REST routes to the supplied router group.
func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	v1.GET("/health", h.health)
	v1.POST("/quotes", h.createQuote)
	v1.GET("/quotes", h.getQuotes)
	v1.GET("/quotes/:id", h.getQuoteByID)
	v1.GET("/quotes/random", h.getRandomQuote)
	v1.PATCH("/quotes/:id", h.updateQuote)
	v1.DELETE("/quotes/:id", h.deleteQuote)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{Data: struct {
		Status string `json:"status"`
	}{Status: "ok"}})
}

func (h *Handler) createQuote(c *gin.Context) {
	var request createQuoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON request body")
		return
	}

	quote, err := h.service.CreateQuote(c.Request.Context(), model.QuoteCreateInput{Text: request.Text, Author: request.Author})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Header("Location", "/v1/quotes/"+quote.ID)
	c.JSON(http.StatusCreated, quoteResponse{Data: toQuotePayload(quote)})
}

func (h *Handler) getQuotes(c *gin.Context) {
	var request getQuotesRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", "Invalid query parameters")
		return
	}

	result, err := h.service.ListQuotes(c.Request.Context(), service.QuoteListInput{
		Author: request.Author,
		Limit:  request.Limit,
		Offset: request.Offset,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	data := make([]quotePayload, 0, len(result.Quotes))
	for _, quote := range result.Quotes {
		data = append(data, toQuotePayload(quote))
	}
	c.JSON(http.StatusOK, quoteListResponse{Data: data, Meta: quoteListMeta{
		Count: result.Count, Limit: result.Limit, Offset: result.Offset,
	}})
}

func (h *Handler) getQuoteByID(c *gin.Context) {
	quote, err := h.service.GetQuoteByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, quoteResponse{Data: toQuotePayload(quote)})
}

func (h *Handler) getRandomQuote(c *gin.Context) {
	quote, err := h.service.GetRandomQuote(c.Request.Context())
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, quoteResponse{Data: toQuotePayload(quote)})
}

func (h *Handler) updateQuote(c *gin.Context) {
	var request updateQuoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	quote, err := h.service.UpdateQuote(c.Request.Context(), model.QuoteUpdateInput{
		ID: c.Param("id"), Text: request.Text, Author: request.Author,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, quoteResponse{Data: toQuotePayload(quote)})
}

func (h *Handler) deleteQuote(c *gin.Context) {
	if err := h.service.DeleteQuote(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidQuoteListOptions):
		writeError(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", "Invalid query parameters")
	case errors.Is(err, model.ErrInvalidQuoteID):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_ID", "Invalid quote ID")
	case errors.Is(err, model.ErrInvalidQuoteText):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_TEXT", "Invalid quote text")
	case errors.Is(err, model.ErrInvalidQuoteAuthor):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_AUTHOR", "Invalid quote author")
	case errors.Is(err, model.ErrNoFieldsToUpdate):
		writeError(c, http.StatusBadRequest, "NO_FIELDS_TO_UPDATE", "No fields to update")
	case errors.Is(err, model.ErrQuoteAlreadyExists):
		writeError(c, http.StatusConflict, "QUOTE_ALREADY_EXISTS", "Quote already exists")
	case errors.Is(err, model.ErrQuoteNotFound):
		writeError(c, http.StatusNotFound, "QUOTE_NOT_FOUND", "Quote not found")
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{Error: apiError{Code: code, Message: message}})
}

func toQuotePayload(quote model.Quote) quotePayload {
	return quotePayload{ID: quote.ID, Text: quote.Text, Author: quote.Author, CreatedAt: quote.CreatedAt}
}
