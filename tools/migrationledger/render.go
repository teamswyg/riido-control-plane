package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	if m.RiidoTaskTitle == "" {
		fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	} else {
		fmt.Fprintf(&b, "> Riido task: %s `%s`\n\n", m.RiidoTask, m.RiidoTaskTitle)
	}
	b.WriteString("Executable SSOT: [`control-plane.riido.json`](control-plane.riido.json).\n\n")
	writeLines(&b, m.Intro)
	renderLoop(&b, m.Loop)
	for _, section := range m.Sections {
		renderSection(&b, section)
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSection(b *strings.Builder, item section) {
	fmt.Fprintf(b, "%s %s\n\n", strings.Repeat("#", item.Level), item.Title)
	writeLines(b, item.Body)
}

func writeLines(b *strings.Builder, lines []string) {
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if len(lines) > 0 && lines[len(lines)-1] != "" {
		b.WriteByte('\n')
	}
}
