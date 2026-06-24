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
	return nil
}
