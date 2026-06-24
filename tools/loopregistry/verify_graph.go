package main

import "fmt"

func verifyEvidenceGraph(m manifest, loopIDs map[string]bool) error {
	nodes := map[string]bool{}
	for _, loop := range m.Loops {
		nodes[loop.ID] = true
	}
	for _, claim := range m.Claims {
		nodes[claim.ID] = true
	}
	for _, edge := range m.EvidenceGraph {
		if edge.From == "" || edge.To == "" || edge.Relation == "" {
			return fmt.Errorf("evidence graph edges must be complete")
		}
		if !nodes[edge.From] || (!nodes[edge.To] && !loopIDs[edge.To]) {
			return fmt.Errorf("evidence graph edge references unknown node: %+v", edge)
		}
	}
	return verifyHarnessPromotionGraphEdges(m)
}

func verifyHarnessPromotionGraphEdges(m manifest) error {
	edges := promotionGraphEdges(m.EvidenceGraph)
	for _, loop := range m.Loops {
		for _, target := range loop.PromotesTo {
			key := loop.ID + "\x00" + target
			if !edges[key] {
				return fmt.Errorf("harness loop %s promotion target %s must have promotes_failure_to graph edge", loop.ID, target)
			}
		}
	}
	return nil
}

func promotionGraphEdges(edges []graphEdge) map[string]bool {
	out := map[string]bool{}
	for _, edge := range edges {
		if edge.Relation == "promotes_failure_to" {
			out[edge.From+"\x00"+edge.To] = true
		}
	}
	return out
}
