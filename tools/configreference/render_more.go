package main

import (
	"fmt"
	"strings"
)

func renderLoop(b *strings.Builder, loop loopEvidence) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "Observe: %s Hypothesis: %s Execute: %s Evaluate: %s Retrospective: %s\n\n",
		loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective)
}

func renderList(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	fmt.Fprintf(b, "%s.\n\n", strings.Join(values, "; "))
}
