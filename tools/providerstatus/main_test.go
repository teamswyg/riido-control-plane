package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/providerstatus/requirements"
)

func TestProviderStatusBehaviorGolden(t *testing.T) {
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
	if got.Surfaces != 7 || got.RoutingStatuses != 4 || got.DistributionChannels != 5 {
		t.Fatalf("unexpected vocabulary evidence: %+v", got)
	}
	if got.ValidationRules != 10 || got.RoutingRules != 8 || got.AuthorizationRules != 2 {
		t.Fatalf("unexpected rule evidence: %+v", got)
	}
	if got.SourceChecks != 8 || got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("unexpected source/loop evidence: %+v", got)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run providerstatus: %v\n%s", err, body)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing evidence from go run: %v", err)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
