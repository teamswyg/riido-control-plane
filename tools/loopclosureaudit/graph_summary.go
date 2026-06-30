package main

import (
	"fmt"
	"sort"
)

const graphSummaryCheckID = "evidence_graph_chain_summary"

func graphSummaryProofKey(c check) string {
	return c.Kind + ":" + c.ID
}

func verifyGraphSummaryCheck(c check, idx indexes) error {
	if c.ID != graphSummaryCheckID {
		return fmt.Errorf("unknown graph summary check %s", c.ID)
	}
	summary := idx.graphSummary
	if summary.ChainCount == 0 {
		return fmt.Errorf("evidence graph summary must count chains")
	}
	if summary.CompleteChains != summary.ChainCount {
		return fmt.Errorf("evidence graph summary has incomplete chains")
	}
	if summary.NextLoopCount != len(idx.nextLoopSummary) {
		return fmt.Errorf("evidence graph summary next-loop count drift")
	}
	if sumNextLoopChains(idx.nextLoopSummary) != summary.ChainCount {
		return fmt.Errorf("evidence graph next-loop summary must cover all chains")
	}
	return nil
}

func graphSummaryProofSurface(idx indexes) *proofSurface {
	return &proofSurface{
		GraphChainCount:       idx.graphSummary.ChainCount,
		GraphCompleteChains:   idx.graphSummary.CompleteChains,
		GraphClaimBoundChains: idx.graphSummary.ClaimBoundChains,
		GraphUnclaimedChains:  idx.graphSummary.UnclaimedChains,
		GraphNextLoopCount:    idx.graphSummary.NextLoopCount,
		GraphNextLoops:        graphNextLoopNames(idx.nextLoopSummary),
	}
}

func graphNextLoopNames(items []graphNextLoopSummary) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.NextLoop)
	}
	sort.Strings(names)
	return names
}
