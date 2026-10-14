package main

import (
	"fmt"
	"strings"
)

func verifyCandidateItem(m manifest, artifact candidateEvidence, item closedLoopCandidate) (intakeSource, error) {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Observation) == "" {
		return intakeSource{}, fmt.Errorf("candidate item must bind id and observation")
	}
	source, ok := findSourceForCandidate(m.Sources, item)
	if !ok {
		return intakeSource{}, fmt.Errorf("candidate %s targets unknown harness/loop edge", item.ID)
	}
	if err := verifyCandidateSourceRef(item, artifact, source); err != nil {
		return intakeSource{}, err
	}
	if err := verifyCandidatePromotionEdge(item, source); err != nil {
		return intakeSource{}, err
	}
	if err := verifyCandidateRequiredNextArtifacts(item, source.ID); err != nil {
		return intakeSource{}, err
	}
	if err := verifyAdoptionPlan(item); err != nil {
		return intakeSource{}, err
	}
	return source, nil
}

func findSourceForCandidate(sources []intakeSource, item closedLoopCandidate) (intakeSource, bool) {
	var fallback intakeSource
	fallbackOK := false
	for _, source := range sources {
		if source.PromotionTarget != item.PromotionTarget || source.HarnessLoop != item.HarnessLoop {
			continue
		}
		if item.SourceRef.CandidateArtifact != "" &&
			source.CandidateArtifact == item.SourceRef.CandidateArtifact {
			return source, true
		}
		if !fallbackOK {
			fallback = source
			fallbackOK = true
		}
	}
	return fallback, fallbackOK
}
