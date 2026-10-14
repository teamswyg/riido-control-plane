package main

import (
	"fmt"
	"strings"
)

func renderDirection(b *strings.Builder, rules directionRules) {
	b.WriteString("\n## Direction Rules\n\n")
	b.WriteString("Allowed imports:\n\n")
	for _, item := range rules.AllowedImports {
		fmt.Fprintf(b, "- `%s`\n", item)
	}
	b.WriteString("\nForbidden Go imports:\n\n")
	fmt.Fprintf(b, "- %d private or unapproved import patterns; see the executable SSOT manifest.\n", len(rules.ForbiddenGoImports))
}

func renderLinks(b *strings.Builder, links []link) {
	b.WriteString("\n## SSOT Links\n\n")
	for _, item := range links {
		fmt.Fprintf(b, "- %s: [`%s`](%s)\n", item.Name, item.Path, item.Path)
	}
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("\n## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}
