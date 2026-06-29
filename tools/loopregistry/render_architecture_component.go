package main

import (
	"fmt"
	"strings"
)

func renderArchitectureComponents(
	b *strings.Builder,
	components []architectureComponent,
) {
	fmt.Fprintln(b, "### Architecture Components")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Component | Paths | Loop sample | Claim sample | Verifier sample | Evidence sample |")
	fmt.Fprintln(b, "| --- | ---: | --- | --- | --- | --- |")
	for _, component := range components {
		fmt.Fprintf(b, "| `%s` | `%d` | %s | %s | %s | %s |\n",
			component.Component,
			component.PathCount,
			architectureComponentDocSample(component.LoopIDs),
			architectureComponentDocSample(component.ClaimIDs),
			architectureComponentDocSample(component.VerifierCommands),
			architectureComponentDocSample(component.EvidenceChainIDs),
		)
	}
	fmt.Fprintln(b)
}
