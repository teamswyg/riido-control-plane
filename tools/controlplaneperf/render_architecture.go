package main

import (
	"fmt"
	"strings"
)

func renderArchitecture(b *strings.Builder, components []architectureComponent) {
	b.WriteString("## Architecture Components\n\n")
	b.WriteString("| ID | Categories | Pressure | Signals | Evidence |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, component := range components {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			component.ID,
			strings.Join(component.HotPathCategories, ", "),
			strings.Join(component.PressureDimensions, ", "),
			strings.Join(component.ObservabilitySignals, ", "),
			strings.Join(component.EvidenceRefs, ", "))
	}
	b.WriteString("\n")
}
