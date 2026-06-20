package main

import (
	"fmt"
	"strings"
)

func renderSection(b *strings.Builder, title string, values []string) {
	b.WriteString("\n## " + title + "\n\n")
	for _, value := range values {
		fmt.Fprintf(b, "- `%s`\n", value)
	}
}

func renderProfiles(b *strings.Builder, profiles []profile) {
	b.WriteString("## Evidence Profiles\n\n")
	b.WriteString("| Profile | Workflow | Artifact | Focus |\n| --- | --- | --- | --- |\n")
	for _, profile := range profiles {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s |\n", profile.ID, profile.Workflow, profile.EvidenceArtifact, profile.Focus)
	}
	b.WriteString("\n")
}

func join(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "`"+value+"`")
	}
	return strings.Join(quoted, ", ")
}
