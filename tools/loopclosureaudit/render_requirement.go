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
	b.WriteString("| ID | Status | Proofs | Statement |\n")
	b.WriteString("| --- | --- | ---: | --- |\n")
	for _, req := range requirements {
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` | %s |\n",
			req.ID, req.Status, req.ProofCount, req.Statement)
	}
	b.WriteString("\n")
}
