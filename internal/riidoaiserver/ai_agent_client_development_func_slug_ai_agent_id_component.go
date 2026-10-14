package riidoaiserver

import (
	"strings"
	"unicode"
)

func slugAIAgentIDComponent(value string) string {
	var b strings.Builder
	previousDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		isAllowed := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isAllowed {
			b.WriteRune(r)
			previousDash = false
			continue
		}
		if !previousDash {
			b.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
