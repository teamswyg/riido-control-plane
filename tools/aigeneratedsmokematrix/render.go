package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/rendertext"
	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"
)

func renderDoc(m manifest, counts operationCounts) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, requirements.GeneratedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`ai-agent-generated-endpoint-smoke-matrix.riido.json`](ai-agent-generated-endpoint-smoke-matrix.riido.json).\n\n")
	b.WriteString("This reader is generated from the smoke matrix gate manifest. ")
	b.WriteString("OpenAPI and the matrix fixture are the executable inputs.\n\n")
	renderCoverage(&b, m, counts)
	rendertext.Loop(&b, m.Loop.Observation, m.Loop.Hypothesis, m.Loop.Execute, m.Loop.Evaluate, m.Loop.Retrospective)
	rendertext.CodeList(&b, "Non-Goals", m.NonGoals)
	return b.String()
}

func renderCoverage(b *strings.Builder, m manifest, counts operationCounts) {
	b.WriteString("## Coverage\n\n")
	b.WriteString("| Class | Count |\n| --- | ---: |\n")
	fmt.Fprintf(b, "| Generated OpenAPI operations | %d |\n", counts.Total)
	fmt.Fprintf(b, "| v1 operations | %d |\n", counts.V1)
	fmt.Fprintf(b, "| v2 operations | %d |\n", counts.V2)
	fmt.Fprintf(b, "| Required smoke tests | %d |\n", len(m.RequiredEvidenceTests))
	fmt.Fprintf(b, "| Source checks | %d |\n\n", len(m.SourceChecks))
	b.WriteString("- OpenAPI: `" + m.OpenAPI + "`\n")
	b.WriteString("- Smoke matrix: `" + m.SmokeMatrix + "`\n")
	b.WriteString("- Gate: `go test ./internal/riidoaiserver -run 'GeneratedEndpointSmoke|SmokeMatrix' -count=1`\n\n")
}
