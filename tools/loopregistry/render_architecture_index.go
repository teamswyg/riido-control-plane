package main

import (
	"fmt"
	"strings"
)

func renderArchitectureIndex(
	b *strings.Builder,
	claims []claimBinding,
	surfaces []claimSurface,
) {
	index := architectureIndexFor(claims, surfaces)
	fmt.Fprintln(b, "## Architecture Index")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "Machine-readable path bindings are emitted in `architecture_index` evidence.")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Components | Paths | Claim bindings | Verifier commands |")
	fmt.Fprintln(b, "| ---: | ---: | ---: | ---: |")
	fmt.Fprintf(b, "| `%d` | `%d` | `%d` | `%d` |\n",
		index.ComponentCount,
		index.PathCount,
		index.BindingCount,
		index.VerifierCommandCount,
	)
	fmt.Fprintln(b)
	renderArchitectureComponents(b, index.Components)
}
