package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the metrics HTTP adapter SSOT.\n\n")
	b.WriteString("## Endpoint\n\n")
	fmt.Fprintf(&b, "- Route: `%s %s`\n", m.Endpoint.Method, m.Endpoint.Path)
	fmt.Fprintf(&b, "- Authorization: `%s:%s`\n", m.Endpoint.Resource, m.Endpoint.Action)
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n\n", m.EvidenceArtifact)
	renderStatuses(&b, m.RequiredStatuses)
	renderFields(&b, m.RequiredFields)
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderStatuses(b *strings.Builder, statuses []statusContract) {
	b.WriteString("## Required Statuses\n\n| Case | Status |\n| --- | ---: |\n")
	for _, status := range statuses {
		fmt.Fprintf(b, "| `%s` | `%d` |\n", status.Case, status.Status)
	}
	b.WriteString("\n")
}
