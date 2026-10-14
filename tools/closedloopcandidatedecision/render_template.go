package main

import (
	"fmt"
	"strings"
)

func renderDecisionTemplates(b *strings.Builder, templates []decisionTemplate) {
	if len(templates) == 0 {
		return
	}
	fmt.Fprintln(b, "## Decision Templates")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Subject Kind | Disposition | Priority | Owner | Review By | Next Artifact |")
	fmt.Fprintln(b, "| --- | --- | --- | --- | --- | --- |")
	for _, template := range templates {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			template.SubjectKind, template.Disposition, template.Priority,
			template.Owner, template.ReviewBy, template.NextArtifact)
	}
	fmt.Fprintln(b)
}
