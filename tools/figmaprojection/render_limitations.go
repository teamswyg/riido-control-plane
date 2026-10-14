package main

import (
	"fmt"
	"strings"
)

func renderLimitations(b *strings.Builder, p projectionManifest, s sourceManifest) {
	b.WriteString("## Mirrored Supporting Tool Limitations\n\n")
	byID := sourceLimitationsByID(s.SupportingToolLimitations)
	for _, item := range p.ToolLimitations {
		source := byID[item.SourceID]
		fmt.Fprintf(b, "### `%s`\n\n", item.SourceID)
		fmt.Fprintf(b, "- tool: %s\n", source.Tool)
		fmt.Fprintf(b, "- observed: %s\n", source.ObservedResult)
		fmt.Fprintf(b, "- local scope: %s\n", item.LocalScope)
		fmt.Fprintf(b, "- authoritative pages: %s\n", backtickList(item.RequiredAuthoritativePages))
		fmt.Fprintf(b, "- authoritative results: %s\n", backtickList(item.RequiredAuthoritativeResults))
		fmt.Fprintf(b, "- source authoritative results: %s\n", backtickList(source.AuthoritativeResult))
		fmt.Fprintf(b, "- forbidden effects: %s\n\n", backtickList(item.ForbiddenProjectionEffects))
		renderLimitationNote(b, item.SourceID)
	}
}

func renderLimitationNote(b *strings.Builder, id string) {
	switch id {
	case "figma-metadata-page-list-underreports-pages.v1":
		b.WriteString("This keeps `expected_pages`, `non_ui_top_level_inventory`, and `legacy_non_ui_absorptions`; contracts #52 is the limitation-local provenance.\n\n")
	case "figma-headless-file-key-placeholder.v1":
		b.WriteString("The `figma.fileKey=headless` value is a headless runtime placeholder; keep `figma.file_key` and `source_contracts_manifest` from the source manifest.\n\n")
	case "figma-onboarding-page-load-timeout.v1":
		b.WriteString("Keep `Wireframe - 온보딩`, `236:33845`, `236:33847`, six onboarding `riido.*` `API Generated` annotations, `non_ui_top_level_inventory`, `child_count=84`, `known_inventory_count=83`, and do not mark onboarding generated paths unresolved after 120s timeouts.\n\n")
	}
}

func backtickList(values []string) string {
	if len(values) == 0 {
		return "`-`"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "`"+value+"`")
	}
	return strings.Join(out, ", ")
}

func sourceLimitationsByID(items []toolLimitation) map[string]toolLimitation {
	out := map[string]toolLimitation{}
	for _, item := range items {
		out[item.ID] = item
	}
	return out
}

func sourceEntriesByID(items []sourceEntry) map[string]sourceEntry {
	out := map[string]sourceEntry{}
	for _, item := range items {
		out[item.NodeID] = item
	}
	return out
}
