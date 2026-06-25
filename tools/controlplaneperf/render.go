package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, e evidence) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	b.WriteString("Executable SSOT: [`control-plane-performance.riido.json`](control-plane-performance.riido.json).\n\n")
	renderSummary(&b, m, e)
	renderCommands(&b, m)
	renderHotPaths(&b, m.HotPaths)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSummary(b *strings.Builder, m manifest, e evidence) {
	b.WriteString("## Evidence Surface\n\n")
	fmt.Fprintf(b, "- hot paths: `%d`\n", e.HotPathCount)
	fmt.Fprintf(b, "- benchmarks: `%d`\n", e.BenchmarkCount)
	fmt.Fprintf(b, "- concurrency tests: `%d`\n", e.TestCount)
	fmt.Fprintf(b, "- optimization candidates: `%d`\n", e.CandidateCount)
	fmt.Fprintf(b, "- local pressure artifact: `%s`\n", m.LocalPressureArtifact)
	fmt.Fprintf(b, "- local pressure scenarios: `%d`\n", len(m.LocalPressureScenarios))
	fmt.Fprintf(b, "- candidate artifact: `%s`\n\n", m.CandidateArtifact)
}

func renderCommands(b *strings.Builder, m manifest) {
	b.WriteString("## Commands\n\n")
	fmt.Fprintf(b, "- lightweight benchmark: `%s`\n", m.BenchmarkCommand)
	fmt.Fprintf(b, "- local pressure: `%s`\n", m.LocalPressureCommand)
	fmt.Fprintf(b, "- manual pressure: `%s`\n", m.ManualPressureCommand)
	fmt.Fprintf(b, "- race/concurrency: `%s`\n", m.RaceCommand)
	fmt.Fprintf(b, "- loopback pprof: `%s`\n", m.PprofCommand)
	fmt.Fprintf(b, "- live load evidence: `%s`\n\n", m.LiveLoadCommand)
}

func renderHotPaths(b *strings.Builder, paths []hotPath) {
	b.WriteString("## Hot Paths\n\n")
	b.WriteString("| ID | Category | Evidence | Candidate |\n| --- | --- | --- | --- |\n")
	for _, path := range paths {
		evidence := append([]string{}, path.Benchmarks...)
		evidence = append(evidence, path.Tests...)
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s |\n",
			path.ID, path.Category, strings.Join(evidence, ", "), path.Candidate)
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
