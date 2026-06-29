package main

import "testing"

func TestPreCommitBaselineEvidenceCarriesExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	got := newEvidence(manifest{EvidenceTTL: 24}, verifyResult{WorkflowScheduled: true})
	if got.GeneratedAt != "2026-06-24T00:00:00Z" {
		t.Fatalf("generated_at = %q", got.GeneratedAt)
	}
	if got.ExpiresAt != "2026-06-25T00:00:00Z" {
		t.Fatalf("expires_at = %q", got.ExpiresAt)
	}
	if got.EvidenceTTL != 24 || !got.WorkflowScheduled {
		t.Fatalf("freshness evidence = %+v", got)
	}
}
