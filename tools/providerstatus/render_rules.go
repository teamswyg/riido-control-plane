package main

import (
	"fmt"
	"strings"
)

func renderRuleSet(b *strings.Builder, title string, rules []rule) {
	b.WriteString("\n## " + title + "\n\n")
	fmt.Fprintf(b, "- verified rules: %s\n", ruleList(rules))
}

func renderAuth(b *strings.Builder, rules []authRule) {
	b.WriteString("\n## Authorization\n\n| Action | Scope |\n| --- | --- |\n")
	for _, rule := range rules {
		fmt.Fprintf(b, "| `%s` | `%s` |\n", rule.Action, rule.Scope)
	}
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("\n## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}

func ruleList(rules []rule) string {
	values := make([]string, 0, len(rules))
	for _, rule := range rules {
		values = append(values, "`"+rule.ID+"`")
	}
	return strings.Join(values, ", ")
}
