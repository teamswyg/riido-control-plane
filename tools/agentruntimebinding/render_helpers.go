package main

import (
	"fmt"
	"strings"
)

func renderSection(b *strings.Builder, title string, values []string) {
	b.WriteString("\n## " + title + "\n\n")
	for _, value := range values {
		fmt.Fprintf(b, "- `%s`\n", value)
	}
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
