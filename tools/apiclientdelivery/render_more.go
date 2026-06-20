package main

import (
	"fmt"
	"strings"
)

func renderFigma(b *strings.Builder, contexts []figmaContext) {
	b.WriteString("## Figma Consumption Contexts\n\n")
	for _, ctx := range contexts {
		fmt.Fprintf(b, "### %s\n\n", ctx.ID)
		fmt.Fprintf(b, "- node ids: %s\n", codeList(ctx.NodeIDs))
		fmt.Fprintf(b, "- generated paths: %s\n", emptyAwareCodeList(ctx.GeneratedPaths))
		fmt.Fprintf(b, "- rule: %s\n", ctx.Rule)
		fmt.Fprintf(b, "- must not own: %s\n\n", ctx.MustNotOwn)
	}
}

func renderModelCatalog(b *strings.Builder, model modelCatalog) {
	b.WriteString("## Runtime Model Catalog\n\n")
	fmt.Fprintf(b, "- policy: %s\n", model.Policy)
	fmt.Fprintf(b, "- rendering: %s\n", model.Rendering)
	fmt.Fprintf(b, "- fixture rule: %s\n\n", model.FixtureRule)
}

func renderList(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, item := range items {
		fmt.Fprintf(b, "- `%s`\n", item)
	}
	b.WriteString("\n")
}
