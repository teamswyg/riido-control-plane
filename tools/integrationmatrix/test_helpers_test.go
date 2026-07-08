package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tempRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return root
}

func baseManifest() manifest {
	return manifest{
		SchemaVersion: "test.v1",
		ID:            "integration-matrix",
		Title:         "Integration Matrix",
		GeneratedDoc:  "docs/generated.md",
		Workflow:      ".github/workflows/integration-matrix.yml",
		Evidence:      "evidence.json",
		PublicGates: []publicGate{{
			Surface:            "public",
			Verification:       "local check",
			ExternalDependency: "none",
		}},
		PrivateGates: []privateGate{{
			Surface:  "private",
			Owner:    "control-plane",
			Evidence: "local evidence",
		}},
		Loop: evidenceLoop{"observe", "hypothesize", "execute", "evaluate", "retro"},
	}
}

func writeManifest(t *testing.T, root string, m manifest) string {
	t.Helper()
	path := filepath.Join(root, "manifest.json")
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}
