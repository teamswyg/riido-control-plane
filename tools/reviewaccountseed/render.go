package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/requirements"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(requirements.GeneratedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the review account seed evidence sidecar.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Seed SSOT: `%s`\n", m.SeedSSOT)
	fmt.Fprintf(&b, "- Owners: `%s`\n\n", strings.Join(m.OwnerPackages, "`, `"))
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
	case "provision":
		return "token hash only, no raw token"
	case "catalog":
		return fmt.Sprintf("%d visible agents, admin=%t", len(tc.WantVisibleAgents), tc.WantAdmin)
	case "provider-status":
		return fmt.Sprintf("%d providers, %d available", tc.WantProviderCount, tc.WantAvailableCount)
	case "http":
		return fmt.Sprintf("catalog %d, provider %d, poll %d", tc.WantCatalogStatus, tc.WantProviderStatus, tc.WantPollStatus)
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
