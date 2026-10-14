package main

import "testing"

const stagingClientCurrentTaskVisualFollowupEvidence = "docs/30-architecture/evidence/staging-client-current-task-visual-followup-2026-07-03.json"

func TestStagingClientCurrentTaskVisualFollowupEvidence(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "staging_client_current_task_visual_followup_2026_07_03") {
		t.Fatal("missing current staging task visual follow-up measurement")
	}
	if !hasEvidenceRef(check, stagingClientCurrentTaskVisualFollowupEvidence) {
		t.Fatal("missing current staging task visual follow-up evidence ref")
	}
	evidence := loadVisualEvidence(t, stagingClientCurrentTaskVisualFollowupEvidence)
	if evidence.Metrics.Thinking != 0 || evidence.Metrics.QueuedCopy != 0 {
		t.Fatalf("current task follow-up shows active drift: %+v", evidence.Metrics)
	}
	assertEvidenceFileHash(t, evidence.Screenshot.Path, evidence.Screenshot.SHA256)
	if !hasObservationStatus(evidence.Observations, "non_reproduction_baseline") {
		t.Fatal("follow-up must stay marked as non-reproduction baseline")
	}
}
