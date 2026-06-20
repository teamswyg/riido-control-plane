package main

import (
	"fmt"
	"strings"
)

func renderAssertions(b *strings.Builder, assertions []string) {
	fmt.Fprintln(b)
	fmt.Fprintln(b, "## Assertions")
	fmt.Fprintln(b)
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
}

func renderLoop(b *strings.Builder, loop loopRecord) {
	fmt.Fprintln(b)
	fmt.Fprintln(b, "## Loop")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
