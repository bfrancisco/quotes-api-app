package quotes

import (
	"strings"

	"github.com/google/uuid"
)

func isEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}

func isValidUUID(input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}

	_, err := uuid.Parse(trimmed)
	return err == nil
}
