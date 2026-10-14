package main

import (
	"fmt"
	"strings"
)

func renderLineBudgetHotspots(b *strings.Builder, hotspots []lineBudgetHotspot) {
	if len(hotspots) == 0 {
		return
	}
	b.WriteString("### Line Budget Hotspots\n\n")
	b.WriteString("| Directory | Files over target | Max lines | Total over target lines |\n")
	b.WriteString("| --- | ---: | ---: | ---: |\n")
	for _, hotspot := range hotspots {
		fmt.Fprintf(b, "| `%s` | %d | %d | %d |\n",
			hotspot.Path, hotspot.Files, hotspot.MaxLines, hotspot.TotalOver)
	}
	b.WriteString("\n")
}
