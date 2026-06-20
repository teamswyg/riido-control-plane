package main

import (
	"fmt"
	"strings"
)

func codeList(items []string) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("`%s`", item))
	}
	return strings.Join(parts, ", ")
}

func emptyAwareCodeList(items []string) string {
	if len(items) == 0 {
		return "`none`"
	}
	return codeList(items)
}

func renderLoop(b *strings.Builder, loop loopRecord) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "| Step | Evidence |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}
