package main

import (
	"fmt"
	"strings"
)

func renderAnnotations(b *strings.Builder, s sourceManifest) {
	policy := s.AnnotationPolicy
	b.WriteString("## API Generated Annotation Policy\n\n")
	fmt.Fprintf(b, "- category: `%s` / `%s`\n", policy.CategoryID, policy.CategoryLabel)
	for _, line := range policy.LabelFormat {
		fmt.Fprintf(b, "- format: %s\n", line)
	}
	b.WriteString("- transport rule: `text/event-stream` responses are `SSE Stream`, non-stream `GET` operations are `Query`, and non-`GET` operations are `Mutation`.\n")
	fmt.Fprintf(b, "- source coverage entry rule: %s\n\n", policy.Rule)
	renderLiveInspection(b, policy.LiveInspection)
	renderRetiredCategories(b, policy.RetiredCategories)
	renderAnnotationExamples(b, s.APIGeneratedAnnotations)
}

func renderLiveInspection(b *strings.Builder, scan liveInspection) {
	b.WriteString("### Live Inspection Counts\n\n")
	fmt.Fprintf(b, "- observed: `%s`; tool: %s\n", scan.ObservedAt, scan.Tool)
	for _, page := range scan.PageCounts {
		fmt.Fprintf(b, "- `%s` `%s`: riido `%d`, API Generated `%d`, missing kind `%d`, missing background `%d`\n",
			page.PageID, page.PageName, page.RiidoAnnotationCount, page.APIGeneratedCount,
			page.MissingOperationKind, page.MissingBackground)
	}
	fmt.Fprintf(b, "- totals: riido `%d`, API Generated `%d`\n\n", scan.TotalRiidoAnnotations, scan.TotalAPIGeneratedAnnotations)
}

func renderRetiredCategories(b *strings.Builder, categories []retiredCategory) {
	b.WriteString("### Retired Categories\n\n")
	for _, item := range categories {
		fmt.Fprintf(b, "- `%s` / `%s`: retired `%s`, live usage `%d`, observed `%s`; %s\n",
			item.CategoryID, item.CategoryLabel, item.RetirementStatus, item.LiveUsageCount,
			item.ObservedAt, item.ToolLimitation)
	}
	b.WriteByte('\n')
}
