package main

import (
	"fmt"
	"strings"
)

func renderScripts(b *strings.Builder, scripts []scriptSpec) {
	b.WriteString("## Scripts\n\n| Script | Summary | Phrases |\n| --- | --- | ---: |\n")
	for _, script := range scripts {
		fmt.Fprintf(b, "| `%s` | %s | `%d` |\n", script.Path, script.Summary, len(script.Contains))
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop loopRecord) {
	b.WriteString("## Evidence Loop\n\n")
	b.WriteString("| Step | Evidence |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n\n", loop.Retrospective)
}
