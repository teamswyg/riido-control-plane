package main

import "testing"

func TestEvidenceGraphRequiresHarnessPromotionEdges(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.EvidenceGraph = removePromotionEdgeForTest(
		m.EvidenceGraph,
		"ai_agent_load_harness",
		"closed_loop_candidate",
	)
	if err := verifyEvidenceGraph(m, loopIDsForTest(m)); err == nil {
		t.Fatal("expected missing harness promotion graph edge to fail")
	}
}

func removePromotionEdgeForTest(edges []graphEdge, from, to string) []graphEdge {
	out := []graphEdge{}
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Relation == "promotes_failure_to" {
			continue
		}
		out = append(out, edge)
	}
	return out
}

func loopIDsForTest(m manifest) map[string]bool {
	out := map[string]bool{}
	for _, loop := range m.Loops {
		out[loop.ID] = true
	}
	return out
}
