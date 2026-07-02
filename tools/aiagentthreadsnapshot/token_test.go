package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotWritesBlockedEvidenceWhenTokenMissing(t *testing.T) {
	output := filepath.Join(t.TempDir(), "snapshot.json")
	err := runMain([]string{
		"-base-url", "https://example.invalid",
		"-workspace-id", "ws",
		"-task-id", "task",
		"-token-env", "RIIDO_TEST_MISSING_TOKEN",
		"-output", output,
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var got report
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Decision.Status != "blocked_missing_bearer_token" {
		t.Fatalf("decision = %+v", got.Decision)
	}
}
