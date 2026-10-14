package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the web frontend API SSOT.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Owners: `%s`\n", strings.Join(m.OwnerPackages, "`, `"))
	fmt.Fprintf(&b, "- Runtime config: `%s`\n\n", strings.Join(m.RuntimeConfigKeys, "`, `"))
	renderCORSCases(&b, m.CORSCases)
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderCORSCases(b *strings.Builder, cases []corsCase) {
	b.WriteString("## CORS Evidence Cases\n\n")
	b.WriteString("| Name | Route | Origin | HTTP | Allow-Origin |\n")
	b.WriteString("| --- | --- | --- | ---: | --- |\n")
	for _, tc := range cases {
		fmt.Fprintf(b, "| `%s` | `%s %s` | `%s` | `%d` | `%s` |\n",
			tc.Name, tc.Method, tc.Path, tc.Origin, tc.WantStatus, tc.WantAllowOrigin)
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "- Observation: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
