package main

import (
	"fmt"
	"strings"
)

func renderLineBudgetUntrackedHotspots(
	b *strings.Builder,
	hotspots []lineBudgetHotspot,
) {
	b.WriteString("### Line Budget Untracked Hotspots\n\n")
	if len(hotspots) == 0 {
		b.WriteString("None. Every over-budget directory is covered by a hotspot ratchet.\n\n")
		return
	}
	b.WriteString("| Directory | Files over target | Max lines | Total over target lines |\n")
	b.WriteString("| --- | ---: | ---: | ---: |\n")
	for _, hotspot := range hotspots {
		fmt.Fprintf(b, "| `%s` | %d | %d | %d |\n",
			hotspot.Path, hotspot.Files, hotspot.MaxLines, hotspot.TotalOver)
	}
	b.WriteString("\n")
}
