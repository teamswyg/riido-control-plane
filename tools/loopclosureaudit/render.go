package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, e evidence) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	b.WriteString("Executable SSOT: [`loop-closure-audit.riido.json`](loop-closure-audit.riido.json).\n\n")
	renderSummary(&b, m, e)
	renderAssertions(&b, m.Assertions)
	renderRequirements(&b, m.Requirements)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSummary(b *strings.Builder, m manifest, e evidence) {
	b.WriteString("## Evidence Surface\n\n")
	fmt.Fprintf(b, "- requirements: `%d`\n", e.RequirementCount)
	fmt.Fprintf(b, "- checks: `%d`\n", e.CheckCount)
	fmt.Fprintf(b, "- evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(b, "- workflow: `%s`\n\n", m.Workflow)
}

func renderAssertions(b *strings.Builder, assertions []string) {
	b.WriteString("## Assertions\n\n")
	for _, assertion := range assertions {
		fmt.Fprintf(b, "- %s\n", assertion)
	}
	b.WriteString("\n")
}

func renderRequirements(b *strings.Builder, requirements []requirement) {
	b.WriteString("## Requirements\n\n")
	b.WriteString("| ID | Checks | Statement |\n| --- | ---: | --- |\n")
	for _, req := range requirements {
		fmt.Fprintf(b, "| `%s` | `%d` | %s |\n", req.ID, len(req.Checks), req.Statement)
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop loopSpec) {
	b.WriteString("## Loop\n\n")
	fmt.Fprintf(b, "- Observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
