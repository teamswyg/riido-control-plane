package main

import (
	"fmt"
	"strings"
)

func renderPublicStatus(b *strings.Builder, status publicStatus) {
	b.WriteString("## Public QA Status\n\n")
	fmt.Fprintf(b, "- overall: `%s`\n", status.Overall)
	fmt.Fprintf(b, "- visibility: `%s`\n", status.Visibility)
	fmt.Fprintf(b, "- raw logs included: `%t`\n", status.RawLogsIncluded)
	fmt.Fprintf(b, "- secrets included: `%t`\n", status.SecretsIncluded)
	fmt.Fprintf(b, "- endpoint details: `%s`\n", status.EndpointDetails)
	fmt.Fprintf(b, "- P0 cycles: `%d`\n", status.P0CycleCount)
	fmt.Fprintf(b, "- P0 partial cycles: `%d`\n", status.P0PartialCount)
	fmt.Fprintf(b, "- partial checks: `%d`\n", status.PartialCount)
	fmt.Fprintf(b, "- stale partials: `%d`\n", status.StalePartialCount)
	fmt.Fprintf(b, "- closed-loop candidates: `%d`\n", status.ClosedLoopCandidates)
	fmt.Fprintf(b, "- next artifact: `%s`\n\n", status.NextArtifact)
}
