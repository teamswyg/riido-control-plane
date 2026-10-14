package main

import (
	"fmt"
	"strings"
)

func renderManifestInventory(b *strings.Builder, e evidence) {
	b.WriteString("## Manifest Inventory\n\n| Group | Count | Sample paths |\n| --- | ---: | --- |\n")
	if len(e.ManifestInventoryByGroup) == 0 {
		b.WriteString("| None | 0 | - |\n\n")
		return
	}
	for _, group := range e.ManifestInventoryByGroup {
		fmt.Fprintf(b, "| `%s` | %d | %s |\n", group.Group, group.Count, manifestSampleText(e, group.Group))
	}
	b.WriteString("\n")
}

func manifestSampleText(e evidence, group string) string {
	for _, sample := range e.ManifestInventorySamples {
		if sample.Group == group {
			return samplePathText(sample.Paths)
		}
	}
	return "None"
}

func samplePathText(paths []string) string {
	if len(paths) == 0 {
		return "None"
	}
	quoted := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted = append(quoted, "`"+path+"`")
	}
	return strings.Join(quoted, "<br>")
}
