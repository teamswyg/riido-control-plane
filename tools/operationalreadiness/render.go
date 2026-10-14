package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, e evidence) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	b.WriteString("Executable SSOT: [`operational-readiness.riido.json`](operational-readiness.riido.json).\n\n")
	renderSummary(&b, m, e)
	renderCompletion(&b, e.Completion)
	renderPublicStatusGeneratedDoc(&b, e.PublicStatus)
	renderNotionOpenLoop(&b, e.NotionOpenLoop)
	renderChecks(&b, m.Checks)
	renderPartials(&b, e.PartialChecks)
	renderPartialPromotion(&b, e.PartialPromotion)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSummary(b *strings.Builder, m manifest, e evidence) {
	b.WriteString("## Evidence Surface\n\n")
	fmt.Fprintf(b, "- loop registry: `%s`\n", m.LoopRegistry)
	fmt.Fprintf(b, "- checks: `%d`\n", e.CheckCount)
	fmt.Fprintf(b, "- measurements: `%d`\n", e.MeasurementCount)
	fmt.Fprintf(b, "- covered: `%d`\n", e.CoveredCount)
	fmt.Fprintf(b, "- partial: `%d`\n", e.PartialCount)
	fmt.Fprintf(b, "- evidence ttl hours: `%d`\n", readinessEvidenceTTLHours)
	fmt.Fprintf(b, "- required categories: `%d`\n", len(e.RequiredCategories))
	fmt.Fprintf(b, "- missing categories: `%d`\n\n", len(e.MissingCategories))
}

func renderChecks(b *strings.Builder, checks []readinessCheck) {
	b.WriteString("## Release Prep Checks\n\n")
	b.WriteString("| Date | Category | Status | Check | Measurements | Evidence | Next |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, check := range checks {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%d` | `%d` | `%s` |\n",
			check.Date, check.Category, check.Status, check.ID,
			len(check.Measurements), len(check.EvidenceRefs), check.NextArtifact)
	}
	b.WriteString("\n")
}

func renderPartials(b *strings.Builder, partials []partialCheck) {
	b.WriteString("## Partial Evidence Queue\n\n")
	if len(partials) == 0 {
		b.WriteString("No partial checks.\n\n")
		return
	}
	b.WriteString("| Check | Category | Age | Stale | Next Artifact | Next Command |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, check := range partials {
		fmt.Fprintf(b, "| `%s` | `%s` | `%d` | `%t` | `%s` | `%s` |\n",
			check.ID, check.Category, check.AgeDays, check.Stale,
			check.NextArtifact, check.NextCommand)
	}
	b.WriteString("\n")
}

func renderLoop(b *strings.Builder, loop loopSpec) {
	b.WriteString("## Loop\n\n")
	fmt.Fprintf(b, "- Observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- Hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- Execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- Evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- Retrospective: %s\n", loop.Retrospective)
}
