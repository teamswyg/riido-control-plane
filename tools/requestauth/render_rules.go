package main

import (
	"fmt"
	"strings"
)

func renderRuleGroups(b *strings.Builder, groups []ruleGroup) {
	b.WriteString("\n## Rule Groups\n\n| Group | Verified Rules |\n| --- | --- |\n")
	for _, group := range groups {
		fmt.Fprintf(b, "| `%s` | %s |\n", group.ID, codeList(group.Rules))
	}
}

func renderLoop(b *strings.Builder, loop loop) {
	b.WriteString("\n## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}

func hasSurface(items []surface, name string) bool {
	for _, item := range items {
		if item.Name == name && item.Role != "" {
			return true
		}
	}
	return false
}

func hasTransport(items []transport, id string) bool {
	for _, item := range items {
		if item.ID == id && strings.TrimSpace(item.Value) != "" {
			return true
		}
	}
	return false
}
