package main

import (
	"fmt"
	"strings"
)

func renderInspection(b *strings.Builder, method inspectionMethod) {
	b.WriteString("## Inspection Method\n\n")
	fmt.Fprintf(b, "- id: `%s`\n", method.ID)
	fmt.Fprintf(b, "- page registry: `%s`\n", method.PageRegistryExpression)
	fmt.Fprintf(b, "- top-level child count: `%s`\n", method.TopLevelChildCountExpression)
	b.WriteString("- rule: supporting evidence only; passive pages can be lazy/unloaded, so supporting tools must not redefine page-level child counts.\n")
	if strings.TrimSpace(method.Rule) != "" {
		fmt.Fprintf(b, "- mirrored rule: %s\n", method.Rule)
	}
	b.WriteByte('\n')
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "- observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- retrospective: %s\n", loop.Retrospective)
}
