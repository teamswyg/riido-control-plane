package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintln(&b, "Executable SSOT: [`evidence-graph.riido.json`](evidence-graph.riido.json).")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Summary")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- manual refresh: `%s`\n", manualRefreshCommand(m.Workflow))
	fmt.Fprintf(&b, "- evidence artifact: `%s`\n", m.Evidence)
	fmt.Fprintf(&b, "- loop registry: `%s`\n", m.LoopRegistry)
	fmt.Fprintf(&b, "- chains: `%d`\n", result.Chains)
	fmt.Fprintf(&b, "- claim refs: `%d`\n", result.ClaimRefs)
	fmt.Fprintf(&b, "- change refs: `%d`\n", result.ChangeRefs)
	fmt.Fprintf(&b, "- verifier refs: `%d`\n", result.VerifierRefs)
	fmt.Fprintf(&b, "- evidence refs: `%d`\n\n", result.EvidenceRefs)
	renderChains(&b, m.Chains)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderChains(b *strings.Builder, chains []chain) {
	fmt.Fprintln(b, "## Evidence Chains")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Chain | Claims | Changes | Verifiers | Evidence | Next Loop |")
	fmt.Fprintln(b, "| --- | ---: | ---: | ---: | ---: | --- |")
	for _, c := range chains {
		fmt.Fprintf(b, "| `%s` | `%d` | `%d` | `%d` | `%d` | `%s` |\n",
			c.ID, len(c.Claims), len(c.Changes), len(c.Verifiers), len(c.Evidence), c.NextLoop)
	}
	fmt.Fprintln(b)
}
