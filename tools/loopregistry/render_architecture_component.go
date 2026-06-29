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
	fmt.Fprintln(b, "| Component | Paths | Claims | Loops | Verifiers | Evidence chains |")
	fmt.Fprintln(b, "| --- | ---: | ---: | ---: | ---: | ---: |")
	for _, component := range components {
		fmt.Fprintf(b, "| `%s` | `%d` | `%d` | `%d` | `%d` | `%d` |\n",
			component.Component,
			component.PathCount,
			len(component.ClaimIDs),
			len(component.LoopIDs),
			len(component.VerifierCommands),
			len(component.EvidenceChainIDs),
		)
	}
	fmt.Fprintln(b)
}
