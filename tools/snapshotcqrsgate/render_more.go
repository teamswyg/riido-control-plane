package main

import (
	"fmt"
	"strings"
)

func renderMeasurementGate(b *strings.Builder, m manifest) {
	b.WriteString("## Measurement Gate\n\n")
	renderList(b, "Signals", m.MeasurementSignals)
	b.WriteString("## Decision Rules\n\n")
	for _, rule := range m.DecisionRules {
		fmt.Fprintf(b, "- `%s`: if %s, `%s` with threshold `%d%%`.\n", rule.ID, rule.When, rule.Action, rule.ThresholdDropPercent)
	}
	b.WriteByte('\n')
}

func renderCandidateSplit(b *strings.Builder, m manifest) {
	b.WriteString("## Candidate Split\n\n")
	renderList(b, "Command Models", m.CandidateSplit.CommandModels)
	renderList(b, "Query Models", m.CandidateSplit.QueryModels)
	renderList(b, "Forbidden Trace Attributes", m.ForbiddenTraceAttributes)
}

func renderList(b *strings.Builder, title string, values []string) {
	b.WriteString("### " + title + "\n\n")
	for _, value := range values {
		fmt.Fprintf(b, "- `%s`\n", value)
	}
	b.WriteByte('\n')
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", loop.Observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", loop.Hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", loop.Execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", loop.Evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n", loop.Retrospective)
}
