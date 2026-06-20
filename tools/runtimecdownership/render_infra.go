package main

import (
	"fmt"
	"strings"
)

func renderInfraBoundary(b *strings.Builder, m manifest, result verifyResult) {
	b.WriteString("## Infra Boundary\n")
	fmt.Fprintf(b, "`%s` consumes `%d` public awareness paths and keeps private topology evidence outside this repository.\n\n",
		m.Infra.Repo, len(m.Infra.Paths))
	fmt.Fprintf(b, "Topology work unit: `%s`; required output categories: `%d`; forbidden handoff categories: `%d`.\n\n",
		m.InfraTopology.WorkUnit, len(m.InfraTopology.RequiredOutput), len(m.InfraTopology.MustNotConsume))
	fmt.Fprintf(b, "Dependency direction: %s %s\n\n", m.DependencyDirection.TopDown, m.DependencyDirection.BottomUp)
	fmt.Fprintf(b, "Infra evidence links verified: `%d`.\n\n", result.InfraLinks)
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n")
	fmt.Fprintf(b, "- observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- retrospective: %s\n", loop.Retrospective)
}
