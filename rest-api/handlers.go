package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/bfrancisco/quotes-api-app/quotes"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store quotes.Store
}

type createQuoteRequest struct {
	Text   string `json:"text"`
	Author string `json:"author"`
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
	v1.POST("/quotes", h.createQuote)
	v1.GET("/health", h.health)
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

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Data: struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		},
	})
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
