package aiagentrisk

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func assertManifestHeader(t *testing.T, manifest evidenceManifest, doc string) {
	t.Helper()
	if manifest.SchemaVersion != "riido-control-plane-ai-agent-risk-evidence.v1" {
		t.Fatalf("unexpected schema_version %q", manifest.SchemaVersion)
	}
	if manifest.ID != "control-plane-ai-agent-risk-evidence" {
		t.Fatalf("unexpected id %q", manifest.ID)
	}
	if manifest.RiidoTask != "RIID-4964" {
		t.Fatalf("unexpected riido_task %q", manifest.RiidoTask)
	}
	expectedDoc := filepath.Clean(strings.TrimPrefix(humanDocPath, "../../"))
	if filepath.Clean(manifest.HumanDoc) != expectedDoc {
		t.Fatalf("unexpected human_doc %q", manifest.HumanDoc)
	}
	if !strings.Contains(doc, "ai-agent-risk-evidence.riido.json") {
		t.Fatal("human doc must link the executable risk evidence manifest")
	}
}

func assertLocalEvidence(t *testing.T, evidence localEvidence, doc string) {
	t.Helper()
	if evidence.Risk == "" || evidence.Status != "verified" || evidence.Proves == "" {
		t.Fatalf("invalid local evidence %+v", evidence)
	}
	if !slices.Contains(requiredRisks, evidence.Risk) {
		t.Fatalf("unexpected local evidence risk %q", evidence.Risk)
	}
	if !strings.HasPrefix(evidence.Test, "Test") {
		t.Fatalf("local evidence %q must name a Go test", evidence.Risk)
	}
	assertTestFunctionExists(t, evidence.Package, evidence.Test)
	if !strings.Contains(doc, evidence.Test) {
		t.Fatalf("human doc must mention evidence test %s", evidence.Test)
	}
}

func assertExternalEvidence(t *testing.T, evidence externalEvidence, doc string) {
	t.Helper()
	if evidence.Risk == "" || evidence.Status != "verified" || evidence.Proves == "" {
		t.Fatalf("invalid external evidence %+v", evidence)
	}
	if !slices.Contains(requiredRisks, evidence.Risk) {
		t.Fatalf("unexpected external evidence risk %q", evidence.Risk)
	}
	if evidence.Repo != "riido-contracts" {
		t.Fatalf("external evidence repo must stay at repo boundary, got %q", evidence.Repo)
	}
	if strings.Contains(evidence.Test, "/") || strings.Contains(evidence.Test, "internal/") {
		t.Fatalf("external evidence must not reference private package paths: %+v", evidence)
	}
	if !strings.Contains(doc, evidence.Test) {
		t.Fatalf("human doc must mention external evidence test %s", evidence.Test)
	}
}

func assertRemainingBoundary(t *testing.T, boundary remainingBoundary) {
	t.Helper()
	if boundary.ID == "" || boundary.Owner == "" || boundary.Reason == "" {
		t.Fatalf("invalid remaining boundary %+v", boundary)
	}
	if !slices.Contains(requiredBoundaries, boundary.ID) {
		t.Fatalf("unexpected remaining boundary %q", boundary.ID)
	}
}
