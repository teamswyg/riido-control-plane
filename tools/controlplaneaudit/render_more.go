package main

import (
	"fmt"
	"strings"
)

func renderSurfaces(b *strings.Builder, surfaces []surfaceEvidence) {
	b.WriteString("## High-Traffic Surfaces\n\n")
	b.WriteString("| ID | Category | Files | Signals | Candidate |\n| --- | --- | --- | --- | --- |\n")
	for _, surface := range surfaces {
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` | `%s` | %s |\n",
			surface.ID, surface.Category, len(surface.Files),
			signalSummary(surface.Files), surface.Candidate)
	}
	b.WriteString("\n")
}

func signalSummary(files []fileEvidence) string {
	counts := map[string]int{}
	for _, file := range files {
		for signal, count := range file.Signals {
			counts[signal] += count
		}
	}
	parts := make([]string, 0, len(counts))
	for _, signal := range signalPatterns {
		if count := counts[signal]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", signal, count))
		}
	}
	if len(parts) == 0 {
		return "source-bound"
	}
	return strings.Join(parts, ", ")
}

func renderLoop(b *strings.Builder, loop loopSpec) {
	b.WriteString("## Loop\n\n")
	fmt.Fprintf(b, "- Observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
