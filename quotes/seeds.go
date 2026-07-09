package quotes

import "time"

// Random quotes for local testing
func SeedQuotes() []Quote {
	return []Quote{
		{ID: "1", Text: "Keep it simple, stupid.", Author: "Kelly Johnson", CreatedAt: time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "2", Text: "Talk is cheap. Show me the code.", Author: "Linus Torvalds", CreatedAt: time.Date(2024, time.January, 2, 10, 0, 0, 0, time.UTC)},
		{ID: "3", Text: "First, solve the problem. Then, write the code.", Author: "John Johnson", CreatedAt: time.Date(2024, time.January, 3, 10, 0, 0, 0, time.UTC)},
		{ID: "4", Text: "Code is like humor. When you have to explain it, it is bad.", Author: "Cory House", CreatedAt: time.Date(2024, time.January, 4, 10, 0, 0, 0, time.UTC)},
		{ID: "5", Text: "Make it work, make it right, make it fast.", Author: "Kent Beck", CreatedAt: time.Date(2024, time.January, 5, 10, 0, 0, 0, time.UTC)},
		{ID: "6", Text: "Programs must be written for people to read, and only incidentally for machines to execute.", Author: "Harold Abelson", CreatedAt: time.Date(2024, time.January, 6, 10, 0, 0, 0, time.UTC)},
		{ID: "7", Text: "Simplicity is prerequisite for reliability.", Author: "Edsger W. Dijkstra", CreatedAt: time.Date(2024, time.January, 7, 10, 0, 0, 0, time.UTC)},
		{ID: "8", Text: "Before software can be reusable it first has to be usable.", Author: "Ralph Johnson", CreatedAt: time.Date(2024, time.January, 8, 10, 0, 0, 0, time.UTC)},
		{ID: "9", Text: "Fix the cause, not the symptom.", Author: "Steve Maguire", CreatedAt: time.Date(2024, time.January, 9, 10, 0, 0, 0, time.UTC)},
		{ID: "10", Text: "Deleted code is debugged code.", Author: "Jeff Sickel", CreatedAt: time.Date(2024, time.January, 10, 10, 0, 0, 0, time.UTC)},
	}
}
