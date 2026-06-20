package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the CloudWatch EMF shape SSOT.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Owner package: `%s`\n\n", m.OwnerPackage)
	b.WriteString("## Required Dimensions\n\n")
	for _, dimension := range m.RequiredDimensions {
		fmt.Fprintf(&b, "- `%s`\n", dimension)
	}
	b.WriteString("\n## Required Metric Units\n\n")
	b.WriteString("| Metric | Unit |\n| --- | --- |\n")
	for _, unit := range m.RequiredMetricUnit {
		fmt.Fprintf(&b, "| `%s` | `%s` |\n", unit.Name, unit.Unit)
	}
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("\n## Evidence Loop\n\n")
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
