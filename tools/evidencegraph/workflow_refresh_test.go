package main

import "testing"

func TestEvidenceGraphWorkflowPublishesStrictRefreshEvidence(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if got.CadenceMinutes <= 0 || got.CadenceMinutes > evidenceGraphEvidenceTTLHours*60 {
		t.Fatalf("cadence_minutes = %d", got.CadenceMinutes)
	}
	if got.WorkflowFile != "evidence-graph.yml" {
		t.Fatalf("workflow_file = %q", got.WorkflowFile)
	}
	if got.ManualRefresh != "gh workflow run evidence-graph.yml --ref main" {
		t.Fatalf("manual_refresh_command = %q", got.ManualRefresh)
	}
}

func TestEvidenceGraphWorkflowRefreshRejectsSlowCadence(t *testing.T) {
	text := "on:\n  schedule:\n    - cron: \"17 20 * * 1\"\n"
	if _, err := refreshCadenceMinutes(text); err == nil {
		t.Fatal("expected weekly cadence to fail")
	}
}

func TestEvidenceGraphWorkflowRefreshRejectsNonStrictArtifact(t *testing.T) {
	text := "uses: actions/upload-artifact@v7\nwith:\n  name: evidence-graph-evidence\n  if-no-files-found: warn\n"
	if workflowUploadsStrictArtifact(text, "evidence-graph-evidence") {
		t.Fatal("expected non-strict upload to fail")
	}
}
