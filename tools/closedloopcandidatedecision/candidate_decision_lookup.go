package main

import (
	"fmt"
	"strings"
)

func decisionsByID(decisions []decisionRecord) map[string]decisionRecord {
	out := map[string]decisionRecord{}
	for _, decision := range decisions {
		out[decision.CandidateID] = decision
	}
	return out
}

func verifyNoOrphanDecisions(
	decisions []decisionRecord,
	candidates []closedLoopCandidate,
	sourceID string,
) error {
	candidateByID := map[string]bool{}
	for _, item := range candidates {
		candidateByID[item.ID] = true
	}
	sourcePrefix := sourceID + ":"
	for _, decision := range decisions {
		if strings.Contains(decision.CandidateID, ":") &&
			!strings.HasPrefix(decision.CandidateID, sourcePrefix) {
			continue
		}
		if !candidateByID[decision.CandidateID] {
			return fmt.Errorf("decision %s has no matching candidate", decision.CandidateID)
		}
	}
	return nil
}
