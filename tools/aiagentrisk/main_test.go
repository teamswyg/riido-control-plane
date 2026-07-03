package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultManifestVerifies(t *testing.T) {
	if err := run(nil); err != nil {
		t.Fatalf("run default manifest: %v", err)
	}
}

func TestRunAcceptsCheckDoc(t *testing.T) {
	if err := run([]string{"-check-doc"}); err != nil {
		t.Fatalf("run check-doc: %v", err)
	}
}

func TestAIAgentRiskBehaviorGolden(t *testing.T) {
	path := filepath.Join(t.TempDir(), "risk-evidence.json")
	if err := run([]string{"-evidence-out", path}); err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidenceResult
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != evidenceSchemaVersion || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.ID != "control-plane-ai-agent-risk-evidence" {
		t.Fatalf("unexpected evidence id: %s", got.ID)
	}
	if got.LocalEvidence != 12 || got.ExternalEvidence != 2 || got.RemainingBoundary != 2 {
		t.Fatalf("unexpected evidence counts: %+v", got)
	}
	if got.LocalEvidence == 0 || got.ExternalEvidence == 0 || got.RemainingBoundary == 0 {
		t.Fatalf("missing evidence counts: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
	if got.Loop.Observation != "Unresolved AI Agent risks were spread across review notes, tests, and cross-repository boundaries, making closure evidence easy to overstate." {
		t.Fatalf("unexpected loop observation: %s", got.Loop.Observation)
	}
}
