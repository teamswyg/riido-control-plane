package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`runtime-deployment-boundary.riido.json`](runtime-deployment-boundary.riido.json).\n\n")
	b.WriteString("Linked runtime CD SSOT: [`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json).\n\n")
	b.WriteString("This reader is generated from the public/private runtime deployment boundary manifest.\n\n")
	renderCoverage(&b, result)
	renderBoundaries(&b, m.Boundaries)
	renderList(&b, "Boundary Rules", m.Rules)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, result verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Boundaries: `%d`; evidence paths: `%d`; phrase checks: `%d`; rules: `%d`.\n\n",
		result.BoundaryCount, result.EvidencePaths, result.PhraseChecks, result.RuleCount)
}

func renderBoundaries(b *strings.Builder, items []boundary) {
	b.WriteString("## Boundaries\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "### %s\n\n", item.ID)
		fmt.Fprintf(b, "- owner: `%s`\n", item.Owner)
		fmt.Fprintf(b, "- scope: %s\n", item.Scope)
		fmt.Fprintf(b, "- does not own: %s\n", strings.Join(item.DoesNotOwn, "; "))
		fmt.Fprintf(b, "- evidence: %s\n\n", evidenceList(item.Evidence))
	}
}
