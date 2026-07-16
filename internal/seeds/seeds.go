package seeds

import (
	"context"

	"github.com/bfrancisco/quotes-api-app/quotes"
)

// SeedQuotes populates the store with an initial set of quotes for local testing.
func SeedQuotes(store quotes.Store) error {
	inputs := []quotes.QuoteCreateInput{
		{Text: "Keep it simple, stupid.", Author: "Kelly Johnson"},
		{Text: "Talk is cheap. Show me the code.", Author: "Linus Torvalds"},
		{Text: "First, solve the problem. Then, write the code.", Author: "John Johnson"},
		{Text: "Code is like humor. When you have to explain it, it is bad.", Author: "Cory House"},
		{Text: "Make it work, make it right, make it fast.", Author: "Kent Beck"},
		{Text: "Programs must be written for people to read, and only incidentally for machines to execute.", Author: "Harold Abelson"},
		{Text: "Simplicity is prerequisite for reliability.", Author: "Edsger W. Dijkstra"},
		{Text: "Before software can be reusable it first has to be usable.", Author: "Ralph Johnson"},
		{Text: "Fix the cause, not the symptom.", Author: "Steve Maguire"},
		{Text: "Deleted code is debugged code.", Author: "Jeff Sickel"},
	}

	for _, input := range inputs {
		if _, err := store.CreateQuote(context.Background(), input); err != nil {
			return err
		}
	}

	return nil
}
