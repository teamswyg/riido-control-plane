package main

import (
	"fmt"
	"strings"
)

func verifyDecisionSourceCoverage(sources []intakeSource, decisions []decisionRecord) error {
	sourceIDs := intakeSourceIDs(sources)
	for _, decision := range decisions {
		sourceID, ok := decisionSourceID(decision)
		if !ok {
			return fmt.Errorf("candidate decision %s must include source id prefix", decision.CandidateID)
		}
		if !sourceIDs[sourceID] {
			return fmt.Errorf("candidate decision %s references unknown intake source %s", decision.CandidateID, sourceID)
		}
	}
	for _, source := range sources {
		if !decisionCoversSource(source, decisions) {
			return fmt.Errorf("intake source %s has no decision seed", source.ID)
		}
	}
	return nil
}

func decisionCoversSource(source intakeSource, decisions []decisionRecord) bool {
	prefix := source.ID + ":"
	for _, decision := range decisions {
		if strings.HasPrefix(decision.CandidateID, prefix) {
			return true
		}
	}
	return false
}

func intakeSourceIDs(sources []intakeSource) map[string]bool {
	out := map[string]bool{}
	for _, source := range sources {
		out[source.ID] = true
	}
	return out
}

func decisionSourceID(decision decisionRecord) (string, bool) {
	sourceID, _, ok := strings.Cut(decision.CandidateID, ":")
	return sourceID, ok && sourceID != ""
}
