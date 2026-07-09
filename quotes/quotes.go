package quotes

import "time"

type Quote struct {
	ID        string
	Text      string
	Author    string
	CreatedAt time.Time
}
