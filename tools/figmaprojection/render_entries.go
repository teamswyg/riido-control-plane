package main

import (
	"fmt"
	"strings"
)

func renderProjectionEntries(b *strings.Builder, p projectionManifest) {
	b.WriteString("## Local Coverage\n\n")
	for _, entry := range p.Entries {
		fmt.Fprintf(b, "- `%s` %s: `%s` / `%s`\n",
			entry.NodeID, entry.Name, entry.ProjectionStatus, entry.SourceCoverageStatus)
		if strings.TrimSpace(entry.LocalScope) != "" {
			fmt.Fprintf(b, "  - scope: %s\n", entry.LocalScope)
		}
		if len(entry.RequiredGeneratedPaths) > 0 {
			fmt.Fprintf(b, "  - generated paths: %s\n", backtickList(entry.RequiredGeneratedPaths))
		}
		if len(entry.ForbiddenGeneratedPathFragments) > 0 {
			fmt.Fprintf(b, "  - forbidden fragments: %s\n", backtickList(entry.ForbiddenGeneratedPathFragments))
		}
		if strings.TrimSpace(entry.NoEndpointReason) != "" {
			fmt.Fprintf(b, "  - no endpoint: %s\n", entry.NoEndpointReason)
		}
	}
	b.WriteByte('\n')
}

func renderAbsorptions(b *strings.Builder, p projectionManifest) {
	renderLegacyAbsorptions(b, p.LegacyAbsorptions)
	renderPlanningAbsorptions(b, p.PlanningAbsorptions)
}

func renderLegacyAbsorptions(b *strings.Builder, items []legacyAbsorption) {
	b.WriteString("## Legacy Non-UI Absorptions\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- `%s` %s absorbed by `%s`; status `%s`\n",
			item.NodeID, item.Name, item.AbsorbedByTopLevelNodeID, item.ProjectionStatus)
		fmt.Fprintf(b, "  - projection: %s\n", item.LocalScope)
		fmt.Fprintf(b, "  - generated paths: %s\n", backtickList(item.RequiredGeneratedPaths))
	}
	b.WriteByte('\n')
}

func renderPlanningAbsorptions(b *strings.Builder, items []planningAbsorption) {
	b.WriteString("## Non-UI Planning Absorptions\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- `%s` %s: `%s`\n", item.NodeID, item.Name, item.ProjectionStatus)
		fmt.Fprintf(b, "  - scope: %s\n", item.LocalScope)
		fmt.Fprintf(b, "  - generated paths: %s\n", backtickList(item.RequiredGeneratedPaths))
		fmt.Fprintf(b, "  - no new endpoint: %s\n", item.NoNewEndpointReason)
	}
	b.WriteByte('\n')
}
