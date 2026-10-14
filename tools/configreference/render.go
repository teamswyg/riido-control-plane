package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`config-reference.riido.json`](config-reference.riido.json).\n\n")
	b.WriteString("This reader is generated from the runtime config manifest and `cmd/riido_ai_server` env reads.\n\n")
	renderCoverage(&b, result)
	renderEntries(&b, m.Entries)
	renderList(&b, "Runtime Rules", m.Rules)
	renderLoop(&b, m.Loop)
	renderList(&b, "Non-Config Facts", m.NonConfigFacts)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, result verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Runtime env reads: `%d`; manifest entries: `%d`; secret/credential entries: `%d`; operator-only entries: `%d`.\n\n",
		result.RuntimeEnvCount, result.ManifestCount, result.SecretCount, result.OperatorCount)
}

func renderEntries(b *strings.Builder, entries []entry) {
	b.WriteString("## Runtime Env\n\n")
	b.WriteString("| Variable | Default | Sensitivity | Meaning |\n| --- | --- | --- | --- |\n")
	for _, entry := range entries {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s |\n",
			entry.Name, entry.Default, entry.Sensitivity, entry.Meaning)
	}
	b.WriteString("\n")
}
