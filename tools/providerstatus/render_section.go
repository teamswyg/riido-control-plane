package main

import (
	"fmt"
	"strings"
)

func renderSection(b *strings.Builder, title string, items []string) {
	b.WriteString("\n## " + title + "\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- `%s`\n", item)
	}
}
