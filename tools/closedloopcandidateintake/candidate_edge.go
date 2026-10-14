package main

import "fmt"

func verifyCandidatePromotionEdge(item closedLoopCandidate, source intakeSource) error {
	if item.PromotionEdge.From != source.HarnessLoop ||
		item.PromotionEdge.To != source.PromotionTarget ||
		item.PromotionEdge.Relation != "promotes_failure_to" {
		return fmt.Errorf("candidate %s promotion_edge does not match intake source", item.ID)
	}
	return nil
}
