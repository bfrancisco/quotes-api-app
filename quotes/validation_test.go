package quotes

import (
	"errors"
	"testing"
	"time"
)

func validationStringPtr(value string) *string {
	return &value
}

func TestQuoteValidate(t *testing.T) {
	t.Parallel()

	validCreatedAt := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		quote   Quote
		wantErr error
	}{
		{
			name: "valid quote",
			quote: Quote{
				ID:        "1",
				Text:      "Simplicity is prerequisite for reliability.",
				Author:    "Edsger W. Dijkstra",
				CreatedAt: validCreatedAt,
			},
		},
		{
			name: "empty id",
			quote: Quote{
				ID:        " ",
				Text:      "Simplicity is prerequisite for reliability.",
				Author:    "Edsger W. Dijkstra",
				CreatedAt: validCreatedAt,
			},
			wantErr: ErrInvalidQuoteID,
		},
		{
			name: "non numeric id",
			quote: Quote{
				ID:        "abc",
				Text:      "Simplicity is prerequisite for reliability.",
				Author:    "Edsger W. Dijkstra",
				CreatedAt: validCreatedAt,
			},
			wantErr: ErrInvalidQuoteID,
		},
		{
			name: "empty text",
			quote: Quote{
				ID:        "1",
				Text:      " ",
				Author:    "Edsger W. Dijkstra",
				CreatedAt: validCreatedAt,
			},
			wantErr: ErrInvalidQuoteText,
		},
		{
			name: "empty author",
			quote: Quote{
				ID:        "1",
				Text:      "Simplicity is prerequisite for reliability.",
				Author:    " ",
				CreatedAt: validCreatedAt,
			},
			wantErr: ErrInvalidQuoteAuthor,
		},
		{
			name: "zero created at",
			quote: Quote{
				ID:     "1",
				Text:   "Simplicity is prerequisite for reliability.",
				Author: "Edsger W. Dijkstra",
			},
			wantErr: ErrInvalidCreateTime,
		},
		{
			name: "future created at",
			quote: Quote{
				ID:        "1",
				Text:      "Simplicity is prerequisite for reliability.",
				Author:    "Edsger W. Dijkstra",
				CreatedAt: time.Now().Add(time.Minute),
			},
			wantErr: ErrInvalidCreateTime,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.quote.Validate()
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestQuoteCreateInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   QuoteCreateInput
		wantErr error
	}{
		{
			name: "valid create input",
			input: QuoteCreateInput{
				Text:   "Simplicity is prerequisite for reliability.",
				Author: "Edsger W. Dijkstra",
			},
		},
		{
			name: "empty text",
			input: QuoteCreateInput{
				Text:   " ",
				Author: "Edsger W. Dijkstra",
			},
			wantErr: ErrInvalidQuoteText,
		},
		{
			name: "empty author",
			input: QuoteCreateInput{
				Text:   "Simplicity is prerequisite for reliability.",
				Author: " ",
			},
			wantErr: ErrInvalidQuoteAuthor,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.input.Validate()
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestQuoteUpdateInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   QuoteUpdateInput
		wantErr error
	}{
		{
			name: "valid update input",
			input: QuoteUpdateInput{
				ID:     "1",
				Text:   validationStringPtr("Simplicity is prerequisite for reliability."),
				Author: validationStringPtr("Edsger W. Dijkstra"),
			},
		},
		{
			name: "valid update input with text only",
			input: QuoteUpdateInput{
				ID:   "1",
				Text: validationStringPtr("Simplicity is prerequisite for reliability."),
			},
		},
		{
			name: "valid update input with author only",
			input: QuoteUpdateInput{
				ID:     "1",
				Author: validationStringPtr("Edsger W. Dijkstra"),
			},
		},
		{
			name: "empty id",
			input: QuoteUpdateInput{
				ID:     " ",
				Text:   validationStringPtr("Simplicity is prerequisite for reliability."),
				Author: validationStringPtr("Edsger W. Dijkstra"),
			},
			wantErr: ErrInvalidQuoteID,
		},
		{
			name: "non numeric id",
			input: QuoteUpdateInput{
				ID:     "abc",
				Text:   validationStringPtr("Simplicity is prerequisite for reliability."),
				Author: validationStringPtr("Edsger W. Dijkstra"),
			},
			wantErr: ErrInvalidQuoteID,
		},
		{
			name: "no fields to update",
			input: QuoteUpdateInput{
				ID: "1",
			},
			wantErr: ErrNoFieldsToUpdate,
		},
		{
			name: "empty text",
			input: QuoteUpdateInput{
				ID:     "1",
				Text:   validationStringPtr(" "),
				Author: validationStringPtr("Edsger W. Dijkstra"),
			},
			wantErr: ErrInvalidQuoteText,
		},
		{
			name: "empty author",
			input: QuoteUpdateInput{
				ID:     "1",
				Text:   validationStringPtr("Simplicity is prerequisite for reliability."),
				Author: validationStringPtr(" "),
			},
			wantErr: ErrInvalidQuoteAuthor,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.input.Validate()
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
