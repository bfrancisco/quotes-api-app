package model

import (
	"errors"
	"testing"
	"time"
)

const validQuoteID = "550e8400-e29b-41d4-a716-446655440000" //valid UUID

func stringPointer(value string) *string {
	return &value
}

// TESTS
// TestQuoteValidation tests the validation of the Quote struct, ensuring that it correctly identifies valid and invalid quotes based on their fields.
// TestInputValidation tests the validation of the QuoteCreateInput and QuoteUpdateInput structs, ensuring that they correctly identify valid and invalid inputs based on their fields.

func TestQuoteValidation(t *testing.T) {
	validQuote := Quote{ID: validQuoteID, Text: "Quote", Author: "Author", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	tests := []struct {
		name    string
		quote   Quote
		wantErr error
	}{
		{"valid", validQuote, nil},
		{"invalid ID", Quote{Text: "Quote", Author: "Author", CreatedAt: validQuote.CreatedAt}, ErrInvalidQuoteID},
		{"invalid text", Quote{ID: validQuoteID, Text: " ", Author: "Author", CreatedAt: validQuote.CreatedAt}, ErrInvalidQuoteText},
		{"invalid author", Quote{ID: validQuoteID, Text: "Quote", Author: " ", CreatedAt: validQuote.CreatedAt}, ErrInvalidQuoteAuthor},
		{"invalid creation time", Quote{ID: validQuoteID, Text: "Quote", Author: "Author"}, ErrInvalidCreateTime},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.quote.Validate(); !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	if err := (QuoteCreateInput{Text: "Quote", Author: "Author"}).Validate(); err != nil {
		t.Fatalf("valid create input error = %v, want nil", err)
	}
	if err := (QuoteCreateInput{Text: " ", Author: "Author"}).Validate(); !errors.Is(err, ErrInvalidQuoteText) {
		t.Fatalf("invalid create error = %v, want %v", err, ErrInvalidQuoteText)
	}
	if err := (QuoteUpdateInput{ID: validQuoteID, Text: stringPointer("Quote")}).Validate(); err != nil {
		t.Fatalf("valid update input error = %v, want nil", err)
	}
	if err := (QuoteUpdateInput{ID: validQuoteID}).Validate(); !errors.Is(err, ErrNoFieldsToUpdate) {
		t.Fatalf("empty update error = %v, want %v", err, ErrNoFieldsToUpdate)
	}
}
