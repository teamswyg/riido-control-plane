package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/agentruntimebinding/requirements"
)

func TestAgentRuntimeBindingBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != requirements.EvidenceSchema || got.ID != requirements.ExpectedID || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.Fields != 5 || got.BindingRules != 6 || got.DeviceRules != 4 || got.SourceChecks != 8 {
		t.Fatalf("unexpected evidence counts: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Evaluate == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
