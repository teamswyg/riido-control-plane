package main

import "testing"

func manifestWithGeneratedCandidateRecord(t *testing.T) manifest {
	t.Helper()
	m := loadDecisionManifestForTest(t)
	m.Decisions = append([]decisionRecord{{
		CandidateID:  "ai-agent-client-testnet-smoke:provider_smoke",
		Disposition:  "triage_required",
		Priority:     "P1",
		Owner:        "agent-platform-loop",
		NextLoop:     "closed_loop_candidate_decision",
		NextArtifact: "claim_binding",
		ReviewBy:     "2026-07-08",
		Reason:       "Test helper binds the generated smoke fixture to an explicit record.",
	}}, m.Decisions...)
	return m
}

func writeGeneratedRecordManifest(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/manifest.json"
	if err := writeJSON(path, manifestWithGeneratedCandidateRecord(t)); err != nil {
		t.Fatal(err)
	}
	return path
}
