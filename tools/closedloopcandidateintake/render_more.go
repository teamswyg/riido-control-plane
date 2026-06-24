package main

import (
	"fmt"
	"strings"
)

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	fmt.Fprintln(b, "## Evidence Loop")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
