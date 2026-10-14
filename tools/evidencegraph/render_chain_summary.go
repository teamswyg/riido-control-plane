package main

import (
	"fmt"
	"strings"
)

func renderChainSummary(
	b *strings.Builder,
	summary chainSummary,
	nextLoops []nextLoopSummary,
) {
	fmt.Fprintln(b, "## Compiled Chain Summary")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- complete chains: `%d` / `%d`\n", summary.CompleteChains, summary.ChainCount)
	fmt.Fprintf(b, "- claim-bound chains: `%d`\n", summary.ClaimBoundChains)
	fmt.Fprintf(b, "- unclaimed chains: `%d`\n", summary.UnclaimedChains)
	fmt.Fprintf(b, "- next-loop targets: `%d`\n\n", summary.NextLoopCount)
	fmt.Fprintln(b, "| Next Loop | Chains | Claims | Changes | Verifiers | Evidence |")
	fmt.Fprintln(b, "| --- | ---: | ---: | ---: | ---: | ---: |")
	for _, item := range nextLoops {
		fmt.Fprintf(b, "| `%s` | `%d` | `%d` | `%d` | `%d` | `%d` |\n",
			item.NextLoop,
			item.Chains,
			item.ClaimRefs,
			item.ChangeRefs,
			item.VerifierRefs,
			item.EvidenceRefs)
	}
	fmt.Fprintln(b)
}
