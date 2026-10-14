package main

func summarizeGraphChains(chains []evidenceChain) graphChainSummary {
	nextLoops := map[string]bool{}
	summary := graphChainSummary{ChainCount: len(chains)}
	for _, chain := range chains {
		if graphChainComplete(chain) {
			summary.CompleteChains++
		}
		if len(chain.Claims) == 0 {
			summary.UnclaimedChains++
		} else {
			summary.ClaimBoundChains++
		}
		nextLoops[chain.NextLoop] = true
	}
	summary.NextLoopCount = len(nextLoops)
	return summary
}

func graphChainComplete(chain evidenceChain) bool {
	return chain.ID != "" &&
		chain.Observation != "" &&
		chain.Hypothesis != "" &&
		len(chain.Changes) > 0 &&
		len(chain.Verifiers) > 0 &&
		len(chain.Evidence) > 0 &&
		chain.Decision != "" &&
		chain.NextLoop != ""
}
