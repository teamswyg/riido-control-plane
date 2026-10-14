package main

import "testing"

func TestEvidenceGraphRequiresClaimEnforcementEdges(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.EvidenceGraph = removeClaimEnforcementEdgeForTest(
		m.EvidenceGraph,
		m.Claims[0].Loop,
		m.Claims[0].ID,
	)
	if err := verifyEvidenceGraph(m, loopIDsForTest(m)); err == nil {
		t.Fatal("expected missing claim enforcement graph edge to fail")
	}
}

func removeClaimEnforcementEdgeForTest(edges []graphEdge, from, to string) []graphEdge {
	out := []graphEdge{}
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Relation == "enforces" {
			continue
		}
		out = append(out, edge)
	}
	return out
}
