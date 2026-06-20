package main

import (
	"fmt"
	"strings"
)

func renderAnnotationExamples(b *strings.Builder, annotations []apiAnnotation) {
	b.WriteString("### Mirrored Annotation Examples\n\n")
	for _, item := range annotations {
		v2 := "v2." + item.CanonicalGeneratedPath
		fmt.Fprintf(b, "- `%s`: `%s` -> `%s`; v2 counterpart `%s`; category `%s`\n",
			item.NodeID, item.FigmaGeneratedPath, item.CanonicalGeneratedPath, v2, item.CategoryLabel)
		if strings.TrimSpace(item.FigmaLabel) != "" {
			fmt.Fprintf(b, "  - label: %s\n", item.FigmaLabel)
		}
		if strings.Contains(item.FigmaLabel, "작업중") {
			b.WriteString("  - stale copy: 상세내용은 작업중입니다\n")
		}
	}
	b.WriteByte('\n')
}
