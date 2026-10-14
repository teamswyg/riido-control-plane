package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`go-ci-baseline.riido.json`](go-ci-baseline.riido.json).\n\n")
	fmt.Fprintf(&b, "- workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- evidence artifact: `%s`\n", m.Evidence)
	fmt.Fprintf(&b, "- gates: `%d`\n", result.Gates)
	fmt.Fprintf(&b, "- phrase checks: `%d`\n\n", result.PhraseChecks)
	renderGates(&b, m.Gates)
	renderLoop(&b, m.Loop)
	renderList(&b, "Non-Goals", m.NonGoals)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderGates(b *strings.Builder, gates []gate) {
	b.WriteString("## Gates\n\n")
	b.WriteString("| Gate | Summary | Phrases |\n| --- | --- | ---: |\n")
	for _, gate := range gates {
		fmt.Fprintf(b, "| `%s` | %s | `%d` |\n", gate.ID, gate.Summary, len(gate.Contains))
	}
	b.WriteString("\n")
}
