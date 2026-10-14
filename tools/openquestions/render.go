package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`open-questions.riido.json`](open-questions.riido.json).\n\n")
	b.WriteString("This reader is generated from the control-plane decision queue manifest.\n\n")
	renderCoverage(&b, result)
	renderQuestions(&b, m.Questions)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, result verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Questions: `%d`; open: `%d`; resolved: `%d`.\n\n",
		result.QuestionCount, result.OpenCount, result.ResolvedCount)
}
