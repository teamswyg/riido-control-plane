package main

import "testing"

const stagingCompletionProgressFollowupEvidence = "docs/30-architecture/evidence/staging-completion-progress-followup-2026-07-02.json"

func TestStagingCompletionProgressFollowupStaysPartial(t *testing.T) {
	evidence := loadCompletionProgressBaselineEvidence(t, stagingCompletionProgressFollowupEvidence)
	if !evidence.Redacted || evidence.Decision.Status != "partial" {
		t.Fatalf("followup must be redacted partial evidence: redacted=%v status=%s", evidence.Redacted, evidence.Decision.Status)
	}
	assertNoActiveCompletionProgress(t, evidence)
}
