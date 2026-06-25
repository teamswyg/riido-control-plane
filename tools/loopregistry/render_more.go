package main

import (
	"fmt"
	"strings"
)

func renderClaims(b *strings.Builder, claims []claimBinding) {
	fmt.Fprintln(b, "## Claim Bindings")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Claim | Loop | Files | Verifiers | Semantic Hash |")
	fmt.Fprintln(b, "| --- | --- | ---: | ---: | --- |")
	for _, claim := range claims {
		hash := claim.SemanticHash
		if len(hash) > 12 {
			hash = hash[:12]
		}
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` | `%d` | `%s` |\n",
			claim.ID, claim.Loop, len(claim.Files), len(claim.Verifiers), hash)
	}
	fmt.Fprintln(b)
}

func renderClaimSurfaces(b *strings.Builder, surfaces []claimSurface) {
	fmt.Fprintln(b, "## Claim Surface Evidence")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Claim | Code | Test | Manifest | Generated docs | Verifies covered | Verifiers | Commands | Evidence chains |")
	fmt.Fprintln(b, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
	for _, surface := range surfaces {
		fmt.Fprintf(b, "| `%s` | `%d` | `%d` | `%d` | `%d` | `%d` | `%d` | `%d` | `%d` |\n",
			surface.ID, len(surface.CodePaths), len(surface.TestPaths),
			len(surface.ManifestPaths), len(surface.GeneratedDocs), len(surface.CoversVerifies), len(surface.Verifiers),
			len(surface.VerifierCommands), len(surface.EvidenceChainIDs))
	}
	fmt.Fprintln(b)
}

func renderGraph(b *strings.Builder, edges []graphEdge) {
	fmt.Fprintln(b, "## Evidence Graph")
	fmt.Fprintln(b)
	for _, edge := range edges {
		fmt.Fprintf(b, "- `%s` --%s--> `%s`\n", edge.From, edge.Relation, edge.To)
	}
	fmt.Fprintln(b)
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	fmt.Fprintln(b, "## Evidence Loop")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "Observe: %s\n\n", loop.Observation)
	fmt.Fprintf(b, "Hypothesis: %s\n\n", loop.Hypothesis)
	fmt.Fprintf(b, "Execute: %s\n\n", loop.Execute)
	fmt.Fprintf(b, "Evaluate: %s\n\n", loop.Evaluate)
	fmt.Fprintf(b, "Retrospective: %s\n", loop.Retrospective)
}
