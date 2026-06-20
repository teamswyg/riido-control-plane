package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`README.riido.json`](README.riido.json).\n\n")
	renderParagraphs(&b, m.Summary)
	renderList(&b, "이 레포가 하는 일", m.Owns)
	renderList(&b, "이 레포가 하지 않는 일", m.DoesNotOwn)
	renderList(&b, "왜 이 작업을 여기서 했나", m.Rationale)
	renderDocLinks(&b, m.DocLinks)
	renderDevelopment(&b, m.Development)
	renderList(&b, "Runtime CD Public Boundary", m.RuntimeCD.Notes)
	renderFlow(&b, m.ContractFlow)
	renderCommands(&b, m.Verification)
	renderList(&b, "Rules", m.Rules)
	renderLoop(&b, m.Loop)
	fmt.Fprintf(&b, "## Module\n\n```text\n%s\n```\n\n", m.ModulePath)
	fmt.Fprintf(&b, "## License\n\n%s.\n", m.License)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderParagraphs(b *strings.Builder, values []string) {
	for _, value := range values {
		b.WriteString(value + "\n\n")
	}
}

func renderList(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	b.WriteString("## " + title + "\n\n")
	for _, value := range values {
		b.WriteString("- " + value + "\n")
	}
	b.WriteByte('\n')
}
