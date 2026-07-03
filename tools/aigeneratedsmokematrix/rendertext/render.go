package rendertext

import (
	"fmt"
	"strings"
)

func Loop(b *strings.Builder, observation, hypothesis, execute, evaluate, retrospective string) {
	b.WriteString("## Evidence Loop\n\n")
	b.WriteString("| Step | Statement |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Observe | %s |\n", observation)
	fmt.Fprintf(b, "| Hypothesis | %s |\n", hypothesis)
	fmt.Fprintf(b, "| Execute | %s |\n", execute)
	fmt.Fprintf(b, "| Evaluate | %s |\n", evaluate)
	fmt.Fprintf(b, "| Retrospective | %s |\n\n", retrospective)
}

func CodeList(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, value := range values {
		fmt.Fprintf(b, "- `%s`\n", value)
	}
}
