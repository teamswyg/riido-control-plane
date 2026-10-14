package main

import (
	"fmt"
	"strings"
)

func writeJSDoc(b *strings.Builder, lines ...string) {
	trimmed := nonEmptyCommentLines(lines)
	if len(trimmed) == 0 {
		return
	}
	b.WriteString("/**\n")
	for _, line := range trimmed {
		fmt.Fprintf(b, " * %s\n", escapeCommentText(line))
	}
	b.WriteString(" */\n")
}

func writeIndentedJSDoc(b *strings.Builder, indent string, lines ...string) {
	trimmed := nonEmptyCommentLines(lines)
	if len(trimmed) == 0 {
		return
	}
	fmt.Fprintf(b, "%s/**\n", indent)
	for _, line := range trimmed {
		fmt.Fprintf(b, "%s * %s\n", indent, escapeCommentText(line))
	}
	fmt.Fprintf(b, "%s */\n", indent)
}

func nonEmptyCommentLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func escapeCommentText(line string) string {
	return strings.ReplaceAll(line, "*/", "* /")
}
