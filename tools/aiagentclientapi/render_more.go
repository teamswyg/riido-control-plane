package main

import (
	"fmt"
	"strings"
)

func renderList(b *strings.Builder, title string, items []string) {
	b.WriteString("\n## " + title + "\n\n")
	fmt.Fprintf(b, "- %s\n", codeList(items))
}

func renderLoop(b *strings.Builder, loop loop) {
	b.WriteString("\n## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}

func codeList(items []string) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, "`"+item+"`")
	}
	return strings.Join(values, ", ")
}
