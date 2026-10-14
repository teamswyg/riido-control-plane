package rendertext

import (
	"fmt"
	"strings"
)

func List(b *strings.Builder, title string, items []string) {
	b.WriteString("\n## " + title + "\n\n")
	fmt.Fprintf(b, "- %s\n", CodeList(items))
}

func CodeList(items []string) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, "`"+item+"`")
	}
	return strings.Join(values, ", ")
}
