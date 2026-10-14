package main

import (
	"fmt"
	"strings"
)

func renderLineBudgetRatchet(b *strings.Builder, result lineBudgetResult) {
	if result.MaxFilesOverTarget <= 0 && result.MaxFileLinesLimit <= 0 {
		return
	}
	b.WriteString("### Line Budget Ratchet\n\n")
	b.WriteString("| Metric | Current | Limit | Slack |\n")
	b.WriteString("| --- | ---: | ---: | ---: |\n")
	fmt.Fprintf(b, "| Files over target | %d | %d | %d |\n",
		result.OverTarget, result.MaxFilesOverTarget, lineBudgetFilesSlack(result))
	fmt.Fprintf(b, "| Max file lines | %d | %d | %d |\n\n",
		result.MaxLines, result.MaxFileLinesLimit, lineBudgetMaxLinesSlack(result))
	b.WriteString("Files over target is reported as surface evidence, but the ratchet fails on max-line or hotspot total-over regressions.\n\n")
}
