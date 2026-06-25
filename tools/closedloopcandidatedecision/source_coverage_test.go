package main

import "testing"

func TestCandidateDecisionRequiresEveryIntakeSource(t *testing.T) {
	root := repoRootForTest(t)
	m := loadDecisionManifestForTest(t)
	m.Decisions = withoutDecisionSource(m.Decisions, "control-plane-performance")
	if _, err := verifyAll(root, m); err == nil {
		t.Fatal("expected missing intake source decision coverage to fail")
	}
}

func withoutDecisionSource(decisions []decisionRecord, sourceID string) []decisionRecord {
	out := make([]decisionRecord, 0, len(decisions))
	prefix := sourceID + ":"
	for _, decision := range decisions {
		if len(decision.CandidateID) >= len(prefix) && decision.CandidateID[:len(prefix)] == prefix {
			continue
		}
		out = append(out, decision)
	}
	return out
}
