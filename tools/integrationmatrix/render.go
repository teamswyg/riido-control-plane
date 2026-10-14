package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`integration-matrix.riido.json`](integration-matrix.riido.json).\n\n")
	b.WriteString("This reader is generated from the public/private verification boundary manifest.\n\n")
	renderCoverage(&b, result)
	renderPublicGates(&b, m.PublicGates)
	renderPrivateGates(&b, m.PrivateGates)
	renderList(&b, "Boundary Rules", m.Rules)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, result verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Public gates: `%d`; PR gates: `%d`; operator-only public gates: `%d`; private gates: `%d`; workflow refs: `%d`; command refs: `%d`.\n\n",
		result.PublicGates, result.PullRequestGates, result.OperatorGates,
		result.PrivateGates, result.WorkflowRefs, result.CommandRefs)
}
