package quotes

import "strings"

func isEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}
