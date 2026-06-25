package main

import "testing"

func TestPressureCandidatesCarryClosedLoopAdoptionPath(t *testing.T) {
	report := pressureReport{Findings: []findingEntry{{
		Candidate: candidateForScenario(scenarios()[0]),
	}}}
	assertPressureCandidatesActionable(t, report)
}

func assertPressureCandidatesActionable(t *testing.T, got pressureReport) {
	t.Helper()
	for _, finding := range got.Findings {
		candidate := finding.Candidate
		if candidate.PromotionTarget != pressurePromotionTarget {
			t.Fatalf("finding candidate missing promotion target: %+v", finding)
		}
		if len(candidate.RequiredNextArtifacts) == 0 || len(candidate.AdoptionPlan) == 0 {
			t.Fatalf("finding candidate missing adoption path: %+v", finding)
		}
	}
}
