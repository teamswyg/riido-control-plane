package main

import (
	"fmt"
	"strings"
)

func renderRiskEvidence(b *strings.Builder, items []riskEvidence) {
	if len(items) == 0 {
		return
	}
	b.WriteString("## AI Agent Risk Evidence\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- `%s`: `%s` proves %s\n", item.Risk, item.Test, item.Proves)
	}
	b.WriteString("\n")
}
