package main

import (
	"fmt"
	"strings"
)

func renderInventory(b *strings.Builder, s sourceManifest) {
	b.WriteString("## API Generated Annotation Inventory\n\n")
	for _, item := range s.APIGeneratedAnnotationInventory {
		v2 := "v2." + item.CanonicalGeneratedPath
		fmt.Fprintf(b, "- %s: `%s` -> `%s`; v2 `%s`; kind `%s`; annotations `%d`\n",
			item.UIArea, item.FigmaGeneratedPath, item.CanonicalGeneratedPath, v2,
			item.OperationKind, item.AnnotationCount)
		fmt.Fprintf(b, "  - background: %s\n", item.Background)
		renderInventorySources(b, item.Sources)
	}
	b.WriteByte('\n')
}

func renderInventorySources(b *strings.Builder, sources []apiSource) {
	for _, source := range sources {
		fmt.Fprintf(b, "  - source `%s` top `%s` entry `%s` nodes %s\n",
			source.PageID, source.TopLevelNodeID, source.CoverageEntryNodeID,
			backtickList(source.NodeIDs))
	}
}
