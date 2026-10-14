package main

import (
	"fmt"
	"strings"
)

func renderFields(b *strings.Builder, fields []string) {
	b.WriteString("## Required JSON Fields\n\n")
	for _, field := range fields {
		fmt.Fprintf(b, "- `%s`\n", field)
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
