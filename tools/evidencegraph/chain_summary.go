package main

type chainSummary struct {
	ChainCount       int `json:"chain_count"`
	CompleteChains   int `json:"complete_chain_count"`
	ClaimBoundChains int `json:"claim_bound_chain_count"`
	UnclaimedChains  int `json:"unclaimed_chain_count"`
	NextLoopCount    int `json:"next_loop_count"`
}

func summarizeChains(chains []chain) chainSummary {
	nextLoops := map[string]bool{}
	summary := chainSummary{ChainCount: len(chains)}
	for _, c := range chains {
		if chainComplete(c) {
			summary.CompleteChains++
		}
		if len(c.Claims) == 0 {
			summary.UnclaimedChains++
		} else {
			summary.ClaimBoundChains++
		}
		nextLoops[c.NextLoop] = true
	}
	summary.NextLoopCount = len(nextLoops)
	return summary
}

func chainComplete(c chain) bool {
	return c.ID != "" &&
		c.Observation != "" &&
		c.Hypothesis != "" &&
		len(c.Changes) > 0 &&
		len(c.Verifiers) > 0 &&
		len(c.Evidence) > 0 &&
		c.Decision != "" &&
		c.NextLoop != ""
}
