package quotes

import "errors"

var (
	ErrInvalidQuoteID     = errors.New("invalid quote ID")
	ErrInvalidQuoteText   = errors.New("invalid quote text")
	ErrInvalidQuoteAuthor = errors.New("invalid quote author")
	ErrInvalidCreateTime  = errors.New("invalid quote creation time")
	ErrQuoteNotFound      = errors.New("quote not found")
	ErrQuoteAlreadyExists = errors.New("quote already exists")
)
