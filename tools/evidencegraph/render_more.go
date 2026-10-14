package main

import (
	"fmt"
	"strings"
)

func renderLoop(b *strings.Builder, loop loopRecord) {
	fmt.Fprintln(b, "## Loop")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- retrospective: %s\n", loop.Retrospective)
}
