package quotes

import (
	"time"
)

func (q Quote) Validate() error {
	if isEmptyString(q.ID) || !isNumeric(q.ID) {
		return ErrInvalidQuoteID
	}

	if isEmptyString(q.Text) {
		return ErrInvalidQuoteText
	}

	if isEmptyString(q.Author) {
		return ErrInvalidQuoteAuthor
	}

	if q.CreatedAt.IsZero() || q.CreatedAt.After(time.Now()) {
		return ErrInvalidCreateTime
	}

	return nil
}

func (q QuoteCreateInput) Validate() error {
	if isEmptyString(q.Text) {
		return ErrInvalidQuoteText
	}

	if isEmptyString(q.Author) {
		return ErrInvalidQuoteAuthor
	}

	return nil
}

func (q QuoteUpdateInput) Validate() error {
	if isEmptyString(q.ID) || !isNumeric(q.ID) {
		return ErrInvalidQuoteID
	}

	if q.Text == nil && q.Author == nil {
		return ErrNoFieldsToUpdate
	}

	if q.Text != nil && isEmptyString(*q.Text) {
		return ErrInvalidQuoteText
	}

	if q.Author != nil && isEmptyString(*q.Author) {
		return ErrInvalidQuoteAuthor
	}

	// Quote existence validation done in store layer.

	return nil
}
