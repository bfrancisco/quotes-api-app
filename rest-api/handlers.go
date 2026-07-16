package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store quotes.Store
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
	Text   *string `json:"text"` // PATCH: using pointer allow distinguishing between "not provided" and "empty string"
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

func NewHandler(store quotes.Store) *Handler {
	return &Handler{store: store}
}

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
	c.JSON(http.StatusOK, healthResponse{
		Data: struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		},
	})
}

func (h *Handler) createQuote(c *gin.Context) {
	var req createQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON request body")
		return
	}

	createdQuote, err := h.store.CreateQuote(c.Request.Context(), quotes.QuoteCreateInput{
		Text:   req.Text,
		Author: req.Author,
	})
	if err != nil {
		h.writeStoreError(c, err)
		return
	}

	c.Header("Location", "/v1/quotes/"+createdQuote.ID)
	c.JSON(http.StatusCreated, quoteResponse{
		Data: toQuotePayload(createdQuote),
	})
}

func (h *Handler) getQuotes(c *gin.Context) {
	var req getQuotesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", "Invalid query parameters")
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Offset < 0 || req.Limit < 1 || req.Limit > 100 {
		writeError(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", "Invalid query parameters")
		return
	}

	var (
		allQuotes []quotes.Quote
		err       error
	)

	if req.Author != "" {
		allQuotes, err = h.store.GetQuotesByAuthor(c.Request.Context(), req.Author)
	} else {
		allQuotes, err = h.store.ListQuotes(c.Request.Context())
	}

	if err != nil {
		h.writeStoreError(c, err)
		return
	}

	start := req.Offset
	if start > len(allQuotes) {
		start = len(allQuotes)
	}

	end := start + req.Limit
	if end > len(allQuotes) {
		end = len(allQuotes)
	}

	page := allQuotes[start:end]

	data := make([]quotePayload, 0, len(page))
	for _, quote := range page {
		data = append(data, toQuotePayload(quote))
	}

	c.JSON(http.StatusOK, quoteListResponse{
		Data: data,
		Meta: quoteListMeta{
			Count:  len(data),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	})
}

func (h *Handler) getQuoteByID(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		h.writeStoreError(c, quotes.ErrInvalidQuoteID)
		return
	}

	quote, err := h.store.GetQuoteByID(c.Request.Context(), id)
	if err != nil {
		h.writeStoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, quoteResponse{
		Data: toQuotePayload(quote),
	})
}

func (h *Handler) getRandomQuote(c *gin.Context) {
	randomQuote, err := h.store.GetRandomQuote(c.Request.Context())
	if err != nil {
		h.writeStoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, quoteResponse{
		Data: toQuotePayload(randomQuote),
	})
}

func (h *Handler) updateQuote(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		h.writeStoreError(c, quotes.ErrInvalidQuoteID)
		return
	}

	var req updateQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	updatedQuote, err := h.store.UpdateQuote(c.Request.Context(), quotes.QuoteUpdateInput{
		ID:     id,
		Text:   req.Text,
		Author: req.Author,
	})
	if err != nil {
		h.writeStoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, quoteResponse{
		Data: toQuotePayload(updatedQuote),
	})
}

func (h *Handler) deleteQuote(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		h.writeStoreError(c, quotes.ErrInvalidQuoteID)
		return
	}

	if err := h.store.DeleteQuote(c.Request.Context(), id); err != nil {
		h.writeStoreError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) writeStoreError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, quotes.ErrInvalidQuoteID):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_ID", "Invalid quote ID")
	case errors.Is(err, quotes.ErrInvalidQuoteText):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_TEXT", "Invalid quote text")
	case errors.Is(err, quotes.ErrInvalidQuoteAuthor):
		writeError(c, http.StatusBadRequest, "INVALID_QUOTE_AUTHOR", "Invalid quote author")
	case errors.Is(err, quotes.ErrNoFieldsToUpdate):
		writeError(c, http.StatusBadRequest, "NO_FIELDS_TO_UPDATE", "No fields to update")
	case errors.Is(err, quotes.ErrQuoteAlreadyExists):
		writeError(c, http.StatusConflict, "QUOTE_ALREADY_EXISTS", "Quote already exists")
	case errors.Is(err, quotes.ErrQuoteNotFound):
		writeError(c, http.StatusNotFound, "QUOTE_NOT_FOUND", "Quote not found")
	default:
		// if happens, this is a bug in the store implementation. We should log it and return a 500.
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func toQuotePayload(quote quotes.Quote) quotePayload {
	return quotePayload{
		ID:        quote.ID,
		Text:      quote.Text,
		Author:    quote.Author,
		CreatedAt: quote.CreatedAt,
	}
}

func isValidUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}
