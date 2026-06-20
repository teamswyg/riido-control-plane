package main

import (
	"fmt"
	"strings"
)

func renderLineBudget(b *strings.Builder, result lineBudgetResult) {
	if result.Target <= 0 {
		return
	}
	fmt.Fprintf(b, "File line budget target: `%d`; files over target: `%d`; max file lines: `%d`.\n\n",
		result.Target, result.OverTarget, result.MaxLines)
	renderLineBudgetRatchet(b, result)
	if len(result.Samples) == 0 {
		return
	}
	b.WriteString("| File | Lines |\n| --- | ---: |\n")
	for _, sample := range result.Samples {
		fmt.Fprintf(b, "| `%s` | %d |\n", sample.Path, sample.Lines)
	}
	b.WriteString("\n")
	renderLineBudgetHotspots(b, result.Hotspots)
}
