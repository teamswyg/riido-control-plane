package main

import (
	"fmt"
	"strings"
)

func renderBoundaries(b *strings.Builder, s sourceManifest) {
	b.WriteString("## Important Boundaries\n\n")
	renderRuntimeEndpointBoundary(b, s)
	b.WriteString("- Web onboarding remains auth/team/product/distribution work; no AI Agent waitlist, marketing, or consent helpers are projected from Figma evidence alone.\n")
	b.WriteString("- Provider install cards remain client/product presentation until a separate SSOT adds a server operation.\n")
	b.WriteString("- Task/subtask assignment stays task-scoped; no project, milestone, intake, mention, or property-filler helpers are shipped without a new SSOT.\n\n")
}

func renderRuntimeEndpointBoundary(b *strings.Builder, s sourceManifest) {
	entry, ok := sourceEntriesByID(s.Entries)["162:23090"]
	if !ok {
		return
	}
	for _, fact := range entry.CoveredFacts {
		if strings.Contains(fact, "node-id=129:17930") {
			fmt.Fprintf(b, "- %s\n", fact)
		}
	}
	b.WriteString("- Runtime settings endpoint-looking label is not a canonical base URL, generated path, or live host export.\n")
}
