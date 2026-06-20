package main

import (
	"fmt"
	"strings"
)

func renderLineBudgetHotspotRatchets(b *strings.Builder, ratchets []lineBudgetHotspotRatchet) {
	if len(ratchets) == 0 {
		return
	}
	b.WriteString("### Line Budget Hotspot Ratchets\n\n")
	b.WriteString("| Directory | Files | Files limit | Max lines | Max limit | Over-target | Over limit |\n")
	b.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, ratchet := range ratchets {
		fmt.Fprintf(b, "| `%s` | %d | %d | %d | %d | %d | %d |\n",
			ratchet.Path, ratchet.Files, ratchet.MaxFiles, ratchet.MaxLines,
			ratchet.MaxLinesLimit, ratchet.TotalOver, ratchet.MaxTotalOver)
	}
	b.WriteString("\n")
}
