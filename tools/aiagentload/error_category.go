package main

import "strings"

func errorCategory(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "context canceled"):
		return "context_cancelled"
	case strings.Contains(lower, "deadline exceeded"),
		strings.Contains(lower, "client.timeout"):
		return "timeout"
	case strings.Contains(lower, "eof"),
		strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "broken pipe"):
		return "connection_closed"
	case strings.Contains(lower, "no such host"):
		return "dns"
	default:
		return "other"
	}
}
