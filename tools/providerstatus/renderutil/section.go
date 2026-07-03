package renderutil

import (
	"fmt"
	"strings"
)

func Section(b *strings.Builder, title string, items []string) {
	b.WriteString("\n## " + title + "\n\n")
	for _, item := range items {
		fmt.Fprintf(b, "- `%s`\n", item)
	}
}
