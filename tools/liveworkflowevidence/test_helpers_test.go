package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func liveTestManifest() manifest {
	return manifest{
		SchemaVersion: "riido-live-workflow.v1",
		ID:            "fixture",
		Title:         "Fixture",
		GeneratedDoc:  "docs/live.md",
		Workflow:      ".github/workflows/live.yml",
		Workflows: []workflowSpec{{
			ID:               "fixture-workflow",
			Path:             ".github/workflows/live.yml",
			SummaryArtifact:  "live-evidence",
			SummaryPath:      "out/live.json",
			EvidenceTTLHours: 24,
			SensitiveInputs:  []string{"TESTNET_TOKEN"},
			AllowedFields:    []string{"generated_at", "expires_at", "evidence_claims"},
			RequiredPhrases:  []string{"custom-required"},
			EvidenceClaims: []claimSpec{{
				ID:            "claim-a",
				Summary:       "summary",
				SourcePhrases: []string{"source-phrase-a"},
			}},
		}},
		Loop: loopRecord{Observation: "o", Hypothesis: "h", Execute: "x", Evaluate: "e", Retrospective: "r"},
	}
}

func writeLiveRepo(t *testing.T, m manifest, workflow string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	workflowPath := repoPath(root, m.Workflows[0].Path)
	if err := os.MkdirAll(filepath.Dir(workflowPath), 0o755); err != nil {
		t.Fatalf("mkdir workflow: %v", err)
	}
	if err := os.WriteFile(workflowPath, []byte(workflow), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	manifestPath := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(manifestPath, body, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return root
}

func validWorkflowText() string {
	return "go run ./tools/liveworkflowevidence -workflow fixture-workflow -evidence-out out/live.json\n" +
		"name: live-evidence\npath: out/live.json\nactions/upload-artifact@v7\n" +
		"TESTNET_TOKEN\ncustom-required\nsource-phrase-a\n"
}
