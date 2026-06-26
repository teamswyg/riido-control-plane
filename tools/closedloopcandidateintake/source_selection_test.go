package main

import "testing"

func TestCandidateIntakeSelectsSourceByCandidateArtifact(t *testing.T) {
	sourceA := intakeSource{
		ID: "a", HarnessLoop: "loop", PromotionTarget: "closed_loop_candidate",
		CandidateArtifact: "a-candidates",
	}
	sourceB := sourceA
	sourceB.ID = "b"
	sourceB.CandidateArtifact = "b-candidates"
	item := closedLoopCandidate{
		HarnessLoop: "loop", PromotionTarget: "closed_loop_candidate",
		SourceRef: candidateSourceRef{CandidateArtifact: "b-candidates"},
	}
	got, ok := findSourceForCandidate([]intakeSource{sourceA, sourceB}, item)
	if !ok || got.ID != "b" {
		t.Fatalf("source = %+v, ok = %v", got, ok)
	}
}
