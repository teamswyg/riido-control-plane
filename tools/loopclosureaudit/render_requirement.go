package main

import (
	"fmt"
	"strings"
)

func renderRequirementProofs(
	b *strings.Builder,
	requirements []requirementEvidence,
) {
	b.WriteString("## Requirements\n\n")
	b.WriteString("| ID | Status | Proofs | Surfaces | Statement |\n")
	b.WriteString("| --- | --- | ---: | ---: | --- |\n")
	for _, req := range requirements {
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` | `%d` | %s |\n",
			req.ID, req.Status, req.ProofCount,
			req.ProofSurfaceCount, req.Statement)
	}
	b.WriteString("\n")
}
