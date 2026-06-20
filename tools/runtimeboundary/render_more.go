package main

import (
	"fmt"
	"strings"
)

func evidenceList(items []evidenceCheck) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("`%s`", item.Path))
	}
	return strings.Join(parts, ", ")
}

func renderList(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop loopRecord) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "| Step | Evidence |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}
