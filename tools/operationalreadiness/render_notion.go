package main

import (
	"fmt"
	"strings"
)

func renderNotionOpenLoop(b *strings.Builder, notion notionEvidence) {
	if notion.CycleCount == 0 {
		return
	}
	b.WriteString("## Notion Open Loop Backfill\n\n")
	fmt.Fprintf(b, "- source: [%s](%s)\n", notion.PageTitle, notion.PageURL)
	fmt.Fprintf(b, "- captured at: `%s`\n", notion.CapturedAt)
	fmt.Fprintf(b, "- cadence hours: `%d`\n", notion.CadenceHours)
	fmt.Fprintf(b, "- cycles: `%d`\n", notion.CycleCount)
	fmt.Fprintf(b, "- p0 cycles: `%d`\n", notion.P0Count)
	fmt.Fprintf(b, "- partial cycles: `%d`\n\n", notion.PartialCount)
	b.WriteString("| Cycle | Status | Codex | Backfilled Check | Next |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, cycle := range notion.Cycles {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			cycle.ID, cycle.Status, cycle.CodexStatus,
			cycle.BackfilledCheck, cycle.RequiredNextArtifact)
	}
	b.WriteString("\n")
}
