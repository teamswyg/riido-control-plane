package main

import "fmt"

func verifyCandidatePromotionEdge(item closedLoopCandidate) error {
	if item.PromotionEdge.From != item.HarnessLoop ||
		item.PromotionEdge.To != item.PromotionTarget ||
		item.PromotionEdge.Relation != "promotes_failure_to" {
		return fmt.Errorf("candidate %s promotion_edge does not match harness target", item.ID)
	}
	return nil
}
