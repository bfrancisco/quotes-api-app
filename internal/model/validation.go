package model

import (
	"time"

	"github.com/bfrancisco/quotes-api-app/internal/helpers"
)

func (q Quote) Validate() error {
	if !helpers.IsValidUUID(q.ID) {
		return ErrInvalidQuoteID
	}

	if helpers.IsEmptyString(q.Text) {
		return ErrInvalidQuoteText
	}

	if helpers.IsEmptyString(q.Author) {
		return ErrInvalidQuoteAuthor
	}

	if q.CreatedAt.IsZero() || q.CreatedAt.After(time.Now()) {
		return ErrInvalidCreateTime
	}

	return nil
}

func (q QuoteCreateInput) Validate() error {
	if helpers.IsEmptyString(q.Text) {
		return ErrInvalidQuoteText
	}

	if helpers.IsEmptyString(q.Author) {
		return ErrInvalidQuoteAuthor
	}

	return nil
}

func (q QuoteUpdateInput) Validate() error {
	if !helpers.IsValidUUID(q.ID) {
		return ErrInvalidQuoteID
	}

	if q.Text == nil && q.Author == nil {
		return ErrNoFieldsToUpdate
	}

	if q.Text != nil && helpers.IsEmptyString(*q.Text) {
		return ErrInvalidQuoteText
	}

	if q.Author != nil && helpers.IsEmptyString(*q.Author) {
		return ErrInvalidQuoteAuthor
	}

	return nil
}
