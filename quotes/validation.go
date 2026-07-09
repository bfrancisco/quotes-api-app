package quotes

import (
	"time"
)

// Call after Create and Update operations.
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
