package main

import (
	"fmt"
	"strings"
)

func renderPublicStatus(b *strings.Builder, status publicStatus) {
	renderPublicStatusWithFreshness(b, status, true)
}

func renderPublicStatusGeneratedDoc(b *strings.Builder, status publicStatus) {
	renderPublicStatusWithFreshness(b, status, false)
}

func renderPublicStatusWithFreshness(
	b *strings.Builder,
	status publicStatus,
	includeFreshness bool,
) {
	b.WriteString("## Public QA Status\n\n")
	fmt.Fprintf(b, "- overall: `%s`\n", status.Overall)
	fmt.Fprintf(b, "- visibility: `%s`\n", status.Visibility)
	if includeFreshness {
		fmt.Fprintf(b, "- generated at: `%s`\n", status.GeneratedAt)
		fmt.Fprintf(b, "- expires at: `%s`\n", status.ExpiresAt)
		fmt.Fprintf(b, "- evidence ttl hours: `%d`\n", status.EvidenceTTLHours)
		fmt.Fprintf(b, "- source workflow: `%s`\n", status.SourceWorkflow)
		fmt.Fprintf(b, "- source commit: `%s`\n", status.SourceCommit)
		fmt.Fprintf(b, "- source run id: `%s`\n", status.SourceRunID)
		if status.SourceRunURL != "" {
			fmt.Fprintf(b, "- source run: `%s`\n", status.SourceRunURL)
		}
	}
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

func renderPublicStatusDoc(status publicStatus) string {
	var b strings.Builder
	renderPublicStatus(&b, status)
	return strings.TrimRight(b.String(), "\n") + "\n"
}
