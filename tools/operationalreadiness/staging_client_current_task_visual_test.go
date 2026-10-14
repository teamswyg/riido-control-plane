package main

import (
	"encoding/json"
	"os"
	"testing"
)

const stagingClientCurrentTaskVisualEvidence = "docs/30-architecture/evidence/staging-client-current-task-visual-2026-07-03.json"

func TestStagingClientCurrentTaskVisualEvidence(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "staging_client_current_task_visual_2026_07_03") {
		t.Fatal("missing current staging task visual measurement")
	}
	if !hasEvidenceRef(check, stagingClientCurrentTaskVisualEvidence) {
		t.Fatal("missing current staging task visual evidence ref")
	}
	evidence := loadCurrentTaskVisualEvidence(t)
	if !evidence.Redacted {
		t.Fatal("current task visual evidence must stay redacted")
	}
	if evidence.NextArtifact != "reproduced_completion_progress_same_conversation_network_snapshot" {
		t.Fatalf("next artifact = %q", evidence.NextArtifact)
	}
	if evidence.Metrics.Thinking != 0 || evidence.Metrics.QueuedCopy != 0 {
		t.Fatalf("current task visual evidence shows active drift: %+v", evidence.Metrics)
	}
	assertEvidenceFileHash(t, evidence.Screenshot.Path, evidence.Screenshot.SHA256)
	if !hasObservationStatus(evidence.Observations, "non_reproduction_baseline") {
		t.Fatal("current visual evidence must be explicitly marked as non-reproduction baseline")
	}
}

func loadCurrentTaskVisualEvidence(t *testing.T) currentTaskVisualEvidence {
	t.Helper()
	return loadVisualEvidence(t, stagingClientCurrentTaskVisualEvidence)
}

func loadVisualEvidence(t *testing.T, path string) currentTaskVisualEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence currentTaskVisualEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func hasObservationStatus(observations []currentTaskObservation, status string) bool {
	for _, observation := range observations {
		if observation.Status == status {
			return observation.Decision != ""
		}
	}
	return false
}
