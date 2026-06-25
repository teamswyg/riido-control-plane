package main

import "fmt"

func verifyClaimEnforcementGraphEdges(m manifest) error {
	edges := claimEnforcementGraphEdges(m.EvidenceGraph)
	for _, claim := range m.Claims {
		key := claim.Loop + "\x00" + claim.ID
		if !edges[key] {
			return fmt.Errorf("claim %s must have %s enforces graph edge", claim.ID, claim.Loop)
		}
	}
	return nil
}

func claimEnforcementGraphEdges(edges []graphEdge) map[string]bool {
	out := map[string]bool{}
	for _, edge := range edges {
		if edge.Relation == "enforces" {
			out[edge.From+"\x00"+edge.To] = true
		}
	}
	return out
}
