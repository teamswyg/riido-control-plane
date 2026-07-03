package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/requirements"
)

func TestAgentCatalogRBACBehaviorGolden(t *testing.T) {
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
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.ProfileID != "rbac" || got.Workflow != ".github/workflows/agent-catalog-rbac.yml" {
		t.Fatalf("unexpected evidence profile: %+v", got)
	}
	if got.EvidenceArtifact != "agent-catalog-rbac-evidence" || got.Focus == "" {
		t.Fatalf("unexpected evidence artifact: %+v", got)
	}
	if got.Rules != 5 || got.Scopes != 8 || got.Routes != len(requirements.RequiredRoutes) {
		t.Fatalf("unexpected RBAC counts: %+v", got)
	}
	if got.RequestDTOs != 2 || got.ResponseDTOs != 2 || got.StoreMethods != 4 || got.SourceChecks != 7 {
		t.Fatalf("unexpected surface counts: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
