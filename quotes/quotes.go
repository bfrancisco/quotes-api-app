package quotes

import "time"

type Quote struct {
	ID        string
	Text      string
	Author    string
	CreatedAt time.Time
}

func NewQuote(id, text, author string) Quote {
	return Quote{
		ID:        id,
		Text:      text,
		Author:    author,
		CreatedAt: time.Now(), // CreatedAt always set to time of creation
	}
}
