package quotes

import "strings"

func isEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}

func isNumeric(input string) bool {
	for _, char := range strings.TrimSpace(input) {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
