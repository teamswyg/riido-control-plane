package main

import "sort"

type nextLoopSummary struct {
	NextLoop     string `json:"next_loop"`
	Chains       int    `json:"chain_count"`
	ClaimRefs    int    `json:"claim_ref_count"`
	ChangeRefs   int    `json:"change_ref_count"`
	VerifierRefs int    `json:"verifier_ref_count"`
	EvidenceRefs int    `json:"evidence_ref_count"`
}

func summarizeNextLoops(chains []chain) []nextLoopSummary {
	byLoop := map[string]*nextLoopSummary{}
	for _, c := range chains {
		item := byLoop[c.NextLoop]
		if item == nil {
			item = &nextLoopSummary{NextLoop: c.NextLoop}
			byLoop[c.NextLoop] = item
		}
		item.Chains++
		item.ClaimRefs += len(c.Claims)
		item.ChangeRefs += len(c.Changes)
		item.VerifierRefs += len(c.Verifiers)
		item.EvidenceRefs += len(c.Evidence)
	}
	return sortedNextLoopSummary(byLoop)
}

func sortedNextLoopSummary(items map[string]*nextLoopSummary) []nextLoopSummary {
	out := make([]nextLoopSummary, 0, len(items))
	for _, item := range items {
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Chains != out[j].Chains {
			return out[i].Chains > out[j].Chains
		}
		return out[i].NextLoop < out[j].NextLoop
	})
	return out
}
