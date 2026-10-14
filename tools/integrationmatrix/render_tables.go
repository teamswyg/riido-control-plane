package main

import (
	"fmt"
	"strings"
)

func renderPublicGates(b *strings.Builder, gates []publicGate) {
	b.WriteString("## Public Gates\n\n")
	b.WriteString("| Surface | PR Gate | Evidence |\n| --- | --- | --- |\n")
	for _, gate := range gates {
		fmt.Fprintf(b, "| %s | `%t` | %s |\n", gate.Surface, gate.PullRequestGate, gate.Verification)
	}
	b.WriteString("\n")
}

func renderPrivateGates(b *strings.Builder, gates []privateGate) {
	b.WriteString("## Private Gates\n\n")
	b.WriteString("| Surface | Owner | Evidence |\n| --- | --- | --- |\n")
	for _, gate := range gates {
		fmt.Fprintf(b, "| %s | %s | %s |\n", gate.Surface, gate.Owner, gate.Evidence)
	}
	b.WriteString("\n")
}

func renderList(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n%s.\n\n", title, strings.Join(values, "; "))
}
