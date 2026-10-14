package main

import (
	"fmt"
	"strings"
)

func renderEvidenceKinds(b *strings.Builder, kinds []evidenceKind) {
	fmt.Fprintln(b, "## Evidence Kinds")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Kind | Description |")
	fmt.Fprintln(b, "| --- | --- |")
	for _, kind := range kinds {
		fmt.Fprintf(b, "| `%s` | %s |\n", kind.Kind, kind.Description)
	}
	fmt.Fprintln(b)
}
