package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/requirements"
)

func TestMetricsHTTPAdapterBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
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
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.Endpoint != requirements.ExpectedEndpoint || got.MetricsSchema != requirements.ExpectedMetricsSchema {
		t.Fatalf("unexpected metrics surface: %+v", got)
	}
	if got.AuthorizedStatus != 200 || got.MissingScopeStatus != 403 || got.UnconfiguredStatus != 503 {
		t.Fatalf("unexpected status evidence: %+v", got)
	}
	if got.JSONFieldsVerified != 12 || got.StatusCasesVerified != 3 || got.SourceChecks != 3 {
		t.Fatalf("unexpected verifier counts: %+v", got)
	}
	if got.HTTPBreakdownRows != 1 || got.StoreBreakdownRows != 5 {
		t.Fatalf("unexpected breakdown rows: %+v", got)
	}
	if got.EvidenceArtifact != requirements.EvidenceArtifact || got.Workflow != requirements.Workflow {
		t.Fatalf("unexpected workflow evidence: %+v", got)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run metricshttpadapter: %v\n%s", err, body)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing evidence from go run: %v", err)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
