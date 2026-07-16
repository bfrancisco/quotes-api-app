package helpers

import (
	"strings"

	"github.com/google/uuid"
)

func IsEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}

func IsValidUUID(input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}

	_, err := uuid.Parse(trimmed)
	return err == nil
}
