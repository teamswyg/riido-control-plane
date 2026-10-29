package riidoaiserver

import (
	"strings"
	"time"
)

// normalizedTaskContextTime keeps only a parseable upstream timestamp, as
// RFC 3339 UTC, so read records never carry free-form upstream text.
func normalizedTaskContextTime(raw string) string {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return parsed.UTC().Format(time.RFC3339Nano)
}
