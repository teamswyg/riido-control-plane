package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the store snapshot/file outbox verification sidecar.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Owner: `%s`\n\n", m.OwnerPackage)
	renderCases(&b, m.Cases)
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderCases(b *strings.Builder, cases []caseSpec) {
	b.WriteString("## Evidence Cases\n\n")
	b.WriteString("| Name | Kind | Expected |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, tc := range cases {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` |\n", tc.Name, tc.Kind, caseSummary(tc))
	}
	b.WriteString("\n")
}

func caseSummary(tc caseSpec) string {
	switch tc.Kind {
	case "outbox":
		return fmt.Sprintf("%d records: %s", tc.WantRecords, strings.Join(tc.WantEvents, ", "))
	case "snapshot":
		return fmt.Sprintf("%d history events, state %s", tc.WantHistoryEvents, tc.WantAssignmentState)
	case "outbox-failure":
		return fmt.Sprintf("%d outbox error, %d latency sample", tc.WantOutboxErrors, tc.WantLatencySamples)
	default:
		return "unknown"
	}
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
