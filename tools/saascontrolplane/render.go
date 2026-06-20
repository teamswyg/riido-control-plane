package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido tasks: `%s`\n\n", strings.Join(m.RiidoTasks, "`, `"))
	fmt.Fprintf(&b, "Executable SSOT: [`saas-control-plane.riido.json`](saas-control-plane.riido.json).\n\n")
	b.WriteString("This reader is generated from the coverage-hub manifest. ")
	b.WriteString("Focused workflows and source anchors are the executable evidence.\n\n")
	renderFacts(&b, m)
	renderBoundaries(&b, m.Boundaries)
	renderLoop(&b, m.Loop)
	renderList(&b, "Non-Goals", m.NonGoals)
	return b.String()
}

func renderFacts(b *strings.Builder, m manifest) {
	b.WriteString("## Coverage Facts\n\n")
	b.WriteString("| Fact | Count |\n| --- | ---: |\n")
	fmt.Fprintf(b, "| Boundaries | %d |\n", len(m.Boundaries))
	fmt.Fprintf(b, "| Focused workflows | %d |\n", len(m.FocusedWorkflows))
	fmt.Fprintf(b, "| Source checks | %d |\n", countSourceChecks(m.Boundaries))
	fmt.Fprintf(b, "| Shared contracts | %d |\n\n", len(m.SharedContracts))
	fmt.Fprintf(b, "- owner package: `%s`\n", m.OwnerPackage)
	fmt.Fprintf(b, "- deployment redaction signals: %s\n\n", phraseList(m.GeneratedDoc, m.RequiredPhrases))
}

func renderBoundaries(b *strings.Builder, boundaries []boundary) {
	b.WriteString("## Executable Boundaries\n\n")
	b.WriteString("| Boundary | Workflow | Artifact | Source checks |\n| --- | --- | --- | ---: |\n")
	for _, item := range boundaries {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %d |\n", item.ID, item.Workflow, emptyDash(item.EvidenceArtifact), len(item.SourceChecks))
	}
	b.WriteString("\n")
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func phraseList(generatedDoc string, phrases []phrase) string {
	var values []string
	for _, phrase := range phrases {
		if phrase.File == generatedDoc {
			values = append(values, "`"+phrase.Contains+"`")
		}
	}
	return strings.Join(values, ", ")
}
