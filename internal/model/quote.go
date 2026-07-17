package model

import "time"

type Quote struct {
	ID        string
	Text      string
	Author    string
	CreatedAt time.Time
}

type QuoteCreateInput struct {
	Text   string
	Author string
}

type QuoteUpdateInput struct {
	ID     string
	Text   *string
	Author *string
}
