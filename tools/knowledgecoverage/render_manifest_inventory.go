package main

import (
	"fmt"
	"strings"
)

func renderManifestInventory(b *strings.Builder, e evidence) {
	b.WriteString("## Manifest Inventory\n\n| Group | Count |\n| --- | ---: |\n")
	if len(e.ManifestInventoryByGroup) == 0 {
		b.WriteString("| None | 0 |\n\n")
		return
	}
	for _, group := range e.ManifestInventoryByGroup {
		fmt.Fprintf(b, "| `%s` | %d |\n", group.Group, group.Count)
	}
	b.WriteString("\n")
}
