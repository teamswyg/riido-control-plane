package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the store-safe routing verification sidecar.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Domain SSOT: `%s`\n", m.DomainSSOT)
	fmt.Fprintf(&b, "- Owner: `%s`\n\n", m.OwnerPackage)
	renderCases(&b, m.Cases)
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderCases(b *strings.Builder, cases []routingCase) {
	b.WriteString("## Routing Evidence Cases\n\n")
	b.WriteString("| Name | Runtime | Allowed | Status | Reason/Error |\n")
	b.WriteString("| --- | --- | ---: | --- | --- |\n")
	for _, tc := range cases {
		reason := tc.WantReason
		if tc.WantErrorContains != "" {
			reason = "error contains " + tc.WantErrorContains
		}
		fmt.Fprintf(b, "| `%s` | `%s` | `%t` | `%s` | `%s` |\n",
			tc.Name, strings.TrimSpace(tc.RuntimeProvider), tc.WantAllowed, tc.WantRoutingStatus, reason)
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
