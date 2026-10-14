package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/rendertext"
)

func renderSources(b *strings.Builder, sources []sourceRef) {
	b.WriteString("## Source Manifests\n\n")
	for _, source := range sources {
		fmt.Fprintf(b, "- %s: `%s`\n", source.Name, source.Path)
	}
	b.WriteString("\n")
}

func renderOwners(b *strings.Builder, owners []owner) {
	b.WriteString("## Ownership\n\n")
	for _, item := range owners {
		fmt.Fprintf(b, "- `%s` owns %s; does not own %s.\n", item.Name, item.Owns, item.DoesNotOwn)
	}
	b.WriteString("\n")
}

func renderDelivery(b *strings.Builder, m manifest) {
	b.WriteString("## Release Trigger And Branch\n\n")
	fmt.Fprintf(b, "- workflow: `%s`\n", m.Delivery.Workflow)
	fmt.Fprintf(b, "- package mode: %s\n", m.Delivery.PackageMode)
	fmt.Fprintf(b, "- delivery mode: %s\n", m.Delivery.DeliverMode)
	fmt.Fprintf(b, "- missing credential behavior: %s\n", m.Delivery.IntentionalFailure)
	fmt.Fprintf(b, "- branch source: %s\n", m.Branch.Source)
	fmt.Fprintf(b, "- branch rule: %s\n", m.Branch.Rule)
	fmt.Fprintf(b, "- branch example: `%s`\n", m.Branch.Example)
	fmt.Fprintf(b, "- guard: %s\n\n", m.Branch.SecretGate)
}

func renderGenerator(b *strings.Builder, g generator) {
	b.WriteString("## Generator Boundary\n\n")
	fmt.Fprintf(b, "- React Query generator: `%s`\n", g.ReactQuery)
	fmt.Fprintf(b, "- handoff generator: `%s`\n", g.Handoff)
	fmt.Fprintf(b, "- artifacts: %s\n", rendertext.CodeList(g.Artifacts))
	fmt.Fprintf(b, "- must not own: %s\n\n", strings.Join(g.MustNotOwn, "; "))
}
