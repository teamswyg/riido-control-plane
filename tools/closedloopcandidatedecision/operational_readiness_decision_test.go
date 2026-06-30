package main

import "testing"

func TestOperationalReadinessDecisionUsesSubjectNextArtifact(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/operational-readiness-candidates.json"
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-29T12:01:00Z")
	if err := generateOperationalReadinessCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	candidate, _, err := loadCandidate(out)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	artifactByID := decisionArtifactsByCandidate(result.DecisionArtifacts)
	if _, ok := artifactByID["operational-readiness:otel_xray_client_surface"]; ok {
		t.Fatal("fresh client-surface readiness evidence must not leave a stale decision")
	}
	for _, item := range candidate.Candidates {
		next, err := subjectNextArtifact(item)
		if err != nil {
			t.Fatal(err)
		}
		got := artifactByID[item.ID]
		if next == "" || got.NextArtifact != next {
			t.Fatalf("candidate %s selected %q, want %q", item.ID, got.NextArtifact, next)
		}
		if got.NextCommand == "" || got.NextCommand != commandForArtifact(t, item, next) {
			t.Fatalf("candidate %s command = %q", item.ID, got.NextCommand)
		}
	}
}

func decisionArtifactsByCandidate(items []decisionArtifactEvidence) map[string]decisionArtifactEvidence {
	out := map[string]decisionArtifactEvidence{}
	for _, item := range items {
		out[item.CandidateID] = item
	}
	return out
}

func commandForArtifact(t *testing.T, item closedLoopCandidate, artifact string) string {
	t.Helper()
	command, ok := adoptionCommandFor(item, artifact)
	if !ok {
		t.Fatalf("candidate %s missing command for %s", item.ID, artifact)
	}
	return command
}
