package quotes

import (
	"errors"
	"time"
)

func (q Quote) Validate() error {
	if isEmptyString(q.ID) {
		return errors.New("quote ID cannot be empty")
	}

	if !isNumeric(q.ID) {
		return errors.New("quote ID should be numeric")
	}

	if isEmptyString(q.Text) {
		return errors.New("quote text cannot be empty")
	}

	if isEmptyString(q.Author) {
		return errors.New("quote author cannot be empty")
	}

	if q.CreatedAt.IsZero() {
		return errors.New("quote creation date cannot be empty")
	}

	if q.CreatedAt.After(time.Now()) {
		return errors.New("quote creation date cannot be in the future")
	}

	return nil
}
