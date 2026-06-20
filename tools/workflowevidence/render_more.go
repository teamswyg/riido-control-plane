package main

import (
	"fmt"
	"strings"
)

func renderWorkflowTable(b *strings.Builder, records []workflowRecord) {
	b.WriteString("## Workflow Evidence\n\n")
	b.WriteString("| Workflow | Status | Evidence Out | Artifact | Reason |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, record := range records {
		fmt.Fprintf(
			b,
			"| `%s` | `%s` | `%t` | `%t` | %s |\n",
			record.Path,
			record.Status,
			record.HasEvidenceOut,
			record.UploadsArtifact,
			tableText(recordReason(record)),
		)
	}
	b.WriteString("\n")
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
	b.WriteString("\n")
}

func recordReason(record workflowRecord) string {
	if record.Reason == "" {
		return "-"
	}
	if record.Next == "" {
		return record.Reason
	}
	return record.Reason + " Next: " + record.Next
}

func tableText(text string) string {
	return strings.ReplaceAll(text, "|", "\\|")
}
