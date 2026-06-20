package main

import (
	"fmt"
	"strings"
)

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	values := []string{
		"Observe: " + loop.Observation,
		"Hypothesis: " + loop.Hypothesis,
		"Execute: " + loop.Execute,
		"Evaluate: " + loop.Evaluate,
		"Retrospective: " + loop.Retrospective,
	}
	fmt.Fprintf(b, "## Evidence Loop\n\n%s.\n\n", strings.Join(values, " "))
}
