package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	b.WriteString("# Runtime CD Ownership\n\n")
	b.WriteString(generatedNotice + "\n\n")
	b.WriteString("Executable SSOT: [`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json).\n")
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	renderEvidence(&b, result)
	renderOwnership(&b, m)
	renderStrategies(&b, m)
	renderPublicBoundary(&b, m, result)
	renderInfraBoundary(&b, m, result)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderEvidence(b *strings.Builder, result verifyResult) {
	b.WriteString("## Evidence\n")
	fmt.Fprintf(b, "- strategies verified: `%d`\n", result.Strategies)
	fmt.Fprintf(b, "- public policies verified: `%d`\n", result.PublicPolicies)
	fmt.Fprintf(b, "- public guards verified: `%d`\n", result.PublicGuards)
	fmt.Fprintf(b, "- forbidden export categories verified: `%d`\n", result.ForbiddenItems)
	fmt.Fprintf(b, "- infra links verified: `%d`\n", result.InfraLinks)
	fmt.Fprintf(b, "- loop fields verified: `%d`\n\n", result.LoopFields)
}

func renderOwnership(b *strings.Builder, m manifest) {
	b.WriteString("## Ownership\n")
	fmt.Fprintf(b, "`%s` owns runtime artifact CD for `%s`.\n\n", m.Current.CDOwner, m.Runtime)
	fmt.Fprintf(b, "`%s` owns topology, IAM, drift, Terraform, and operator evidence.\n\n", m.Current.TopologyOwner)
	b.WriteString("Public deploys promote artifacts by preserving the live task-definition shape and replacing only the application image.\n\n")
	b.WriteString("Preserving infra is the same ownership rule for rolling and CodeDeploy deployment paths.\n\n")
}

func renderStrategies(b *strings.Builder, m manifest) {
	b.WriteString("## Current And Future Strategies\n")
	fmt.Fprintf(b, "- current: `%s` (`%s`) via `%s`\n", m.Current.ID, m.Current.Status, m.Current.Workflow)
	for _, mode := range m.OptionalModes {
		fmt.Fprintf(b, "- optional: `%s` (`%s`) via `%s`\n", mode.ID, mode.Status, mode.Workflow)
	}
	for _, future := range m.Future {
		fmt.Fprintf(b, "- future: `%s` (`%s`)\n", future.ID, future.Status)
	}
	b.WriteByte('\n')
}

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
