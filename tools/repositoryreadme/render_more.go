package main

import (
	"fmt"
	"strings"
)

func renderFlow(b *strings.Builder, flow []string) {
	if len(flow) == 0 {
		return
	}
	b.WriteString("## Contract / Generated Client 흐름\n\n```text\n")
	for i, step := range flow {
		if i == 0 {
			b.WriteString(step + "\n")
			continue
		}
		b.WriteString("  -> " + step + "\n")
	}
	b.WriteString("```\n\n")
}

func renderCommands(b *strings.Builder, commands []string) {
	b.WriteString("## 검증\n\n```bash\n")
	for _, command := range commands {
		b.WriteString(command + "\n")
	}
	b.WriteString("```\n\n")
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n\n", loop.Retrospective)
}
