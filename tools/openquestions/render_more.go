package main

import (
	"fmt"
	"strings"
)

func renderQuestions(b *strings.Builder, questions []question) {
	b.WriteString("## Decision Queue\n\n")
	for _, item := range questions {
		fmt.Fprintf(b, "### %s\n\n", item.ID)
		fmt.Fprintf(b, "- status: `%s`\n", item.Status)
		fmt.Fprintf(b, "- area: %s\n", item.Area)
		fmt.Fprintf(b, "- owner: `%s`\n", item.Owner)
		fmt.Fprintf(b, "- question: %s\n", item.Question)
		fmt.Fprintf(b, "- stance: %s\n", item.Stance)
		fmt.Fprintf(b, "- next artifact: %s\n", item.NextArtifact)
		if item.Resolution != "" {
			fmt.Fprintf(b, "- resolution: %s\n", item.Resolution)
		}
		b.WriteString("\n")
	}
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
