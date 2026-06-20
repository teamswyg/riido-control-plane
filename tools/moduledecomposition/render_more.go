package main

import (
	"fmt"
	"strings"
)

func renderPackages(b *strings.Builder, entries []packageEntry) {
	b.WriteString("## Packages\n\n")
	grouped := map[string][]string{}
	order := make([]string, 0)
	for _, entry := range entries {
		if _, ok := grouped[entry.Kind]; !ok {
			order = append(order, entry.Kind)
		}
		grouped[entry.Kind] = append(grouped[entry.Kind], "`"+entry.Path+"`")
	}
	for _, kind := range order {
		fmt.Fprintf(b, "- `%s`: %s\n", kind, strings.Join(grouped[kind], ", "))
	}
	b.WriteString("\n")
}

func renderList(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n%s.\n\n", title, strings.Join(values, "; "))
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	values := []string{
		"Observe: " + loop.Observation,
		"Hypothesis: " + loop.Hypothesis,
		"Execute: " + loop.Execute,
		"Evaluate: " + loop.Evaluate,
		"Retrospective: " + loop.Retrospective,
	}
	fmt.Fprintf(b, "## Evidence Loop\n\n%s.\n\n", strings.Join(values, " "))
}
