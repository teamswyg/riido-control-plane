package main

import "testing"

func TestCandidateDecisionRejectsOrphanDecision(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.Decisions = append(m.Decisions, decisionRecord{
		CandidateID:  "missing-candidate",
		Disposition:  "triage_required",
		Priority:     "P2",
		Owner:        "agent-platform-loop",
		NextLoop:     "closed_loop_candidate_decision",
		NextArtifact: "claim_binding",
		ReviewBy:     "2026-07-01",
		Reason:       "Orphan decision should not survive candidate verification.",
	})
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(root, out); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateDecisions(root, m, out); err == nil {
		t.Fatal("expected orphan decision to fail")
	}
}
