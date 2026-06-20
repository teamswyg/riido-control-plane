package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	fmt.Fprintf(&b, "%s\n\n", generatedNotice)
	fmt.Fprintf(&b, "> Generated from `%s`.\n\n", defaultManifest)
	fmt.Fprintf(&b, "Schema: `%s`\n\n", m.SchemaVersion)
	fmt.Fprintf(&b, "Evidence artifact: `%s`\n\n", m.Evidence)
	renderWorkflowTable(&b, result.Records)
	renderAssertions(&b, m.Assertions)
	renderLoop(&b, m.Loop)
	fmt.Fprintf(&b, "\n## Verification\n\n")
	fmt.Fprintf(&b, "- Workflow count: `%d`\n", result.WorkflowCount)
	fmt.Fprintf(&b, "- Phrase checks: `%d`\n", result.PhraseChecks)
	return b.String()
}

func renderWorkflowTable(b *strings.Builder, records []workflowRecord) {
	fmt.Fprintln(b, "## Redacted Live Workflows")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Workflow | Summary Artifact | Summary Path | Sensitive Inputs |")
	fmt.Fprintln(b, "| --- | --- | --- | --- |")
	for _, record := range records {
		fmt.Fprintf(
			b,
			"| `%s` | `%s` | `%s` | `%s` |\n",
			record.Path,
			record.SummaryArtifact,
			record.SummaryPath,
			strings.Join(record.SensitiveInputs, ", "),
		)
	}
}
