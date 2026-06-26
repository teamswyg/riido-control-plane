package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, e evidence) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	b.WriteString("Executable SSOT: [`control-plane-high-traffic-audit.riido.json`](control-plane-high-traffic-audit.riido.json).\n\n")
	renderSummary(&b, e)
	renderCommands(&b, m)
	renderCategoryCoverage(&b, m.RequiredCategories, e.CategoryCounts)
	renderSurfaces(&b, e.Surfaces)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSummary(b *strings.Builder, e evidence) {
	b.WriteString("## Evidence Surface\n\n")
	fmt.Fprintf(b, "- surfaces: `%d`\n", e.SurfaceCount)
	fmt.Fprintf(b, "- candidates: `%d`\n", e.CandidateCount)
	fmt.Fprintf(b, "- assertions: `%d`\n", e.AssertionCount)
	fmt.Fprintf(b, "- required categories: `%d`\n", len(e.RequiredCategories))
	fmt.Fprintf(b, "- missing categories: `%d`\n", len(e.MissingCategories))
	fmt.Fprintf(b, "- pprof commands: `%d`\n\n", len(e.PprofCommands))
}

func renderCommands(b *strings.Builder, m manifest) {
	b.WriteString("## Commands\n\n")
	fmt.Fprintf(b, "- benchmark: `%s`\n", m.BenchmarkCommand)
	fmt.Fprintf(b, "- local pressure: `%s`\n", m.LocalPressureCommand)
	fmt.Fprintf(b, "- manual pressure: `%s`\n", m.ManualPressureCommand)
	fmt.Fprintf(b, "- local pressure pprof: `%s`\n", m.LocalPprofCommand)
	fmt.Fprintf(b, "- race/concurrency: `%s`\n", m.RaceCommand)
	for _, command := range m.PprofCommands {
		fmt.Fprintf(b, "- pprof: `%s`\n", command)
	}
	b.WriteString("\n")
}
