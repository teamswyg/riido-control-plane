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

func renderArchitectureFileIndex(b *strings.Builder, rows []architectureFileEvidence) {
	b.WriteString("## Architecture File Index\n\n")
	b.WriteString("| File | Components | Hot Paths | Pressure | Signals | Evidence | Target Commands |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | `%d` |\n",
			row.Path,
			strings.Join(row.ComponentIDs, ", "),
			strings.Join(row.HotPathIDs, ", "),
			strings.Join(row.PressureDimensions, ", "),
			strings.Join(row.ObservabilitySignals, ", "),
			strings.Join(row.EvidenceRefs, ", "),
			len(row.TargetVerifierCommands))
	}
	b.WriteString("\n")
}
