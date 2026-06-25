package main

import (
	"fmt"
	"strings"
)

func renderCategoryCoverage(b *strings.Builder, required []string, counts map[string]int) {
	b.WriteString("## Required Category Coverage\n\n")
	b.WriteString("| Category | Surfaces |\n| --- | --- |\n")
	for _, category := range required {
		fmt.Fprintf(b, "| `%s` | `%d` |\n", category, counts[category])
	}
	b.WriteString("\n")
}
